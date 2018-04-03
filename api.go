package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"encoding/json"

	"io"

	"github.com/boltdb/bolt"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"gopkg.in/gin-gonic/gin.v1"
)

func initAPI(engine *gin.Engine) {
	r := engine.Group("/api")
	// API
	r.GET("/link", handleLinkApi)
	r.GET("/new", handleNewApi)
	r.POST("/files", handleFilesApi)
	r.POST("/files/:id/:tag", handleAddTagApi)
	r.DELETE("/files/:id/:tag", handleDeleteTagApi)
	r.POST("/file/:id", handleFileApi)
	r.POST("/search", handleSearchApi)
}

//-----------------------
//-- API Handlers
//-----------------------

func handleLinkApi(c *gin.Context) {
	config.RedirectURL = c.Query("link")
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprint(c.Writer, authURL)
}

func handleNewApi(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		log.Println("No code")
	}
	fmt.Println("got code ....", code)
	var err error
	tok, err = config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Unable to retrieve token from web %v", err)
		return
	}
	fmt.Println("TOK2:", tok)
	tb, err := json.Marshal(tok)
	if err != nil {
		log.Printf("Unable to retrieve token from web %v", err)
	}
	ioutil.WriteFile("tok.json", tb, 0600)
	client = config.Client(ctx, tok)
	service, err = drive.New(client)
	x, err := service.About.Get().Fields("user(permissionId, emailAddress)").Do()
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	id := x.User.PermissionId
	services[id] = service
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}
	idb := []byte(id)
	fmt.Println("Registered a user")
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(idb))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		b.CreateBucket([]byte("files"))

		tk, err := json.Marshal(tok)
		if err != nil {
			return fmt.Errorf("marshal json: %s", err)
		}
		err = b.Put([]byte("token"), tk)
		if err != nil {
			return fmt.Errorf("Put token: %s", err)
		}
		return nil
	})
	c.Writer.Write(idb)
}

func getFiles(user string) map[string]*file {
	filenames := make(map[string]*file)
	var f1 *file
	serv := services[user]
	if serv == nil {
		fmt.Println("FAIL")
		return nil
	}
	i := 0
	serv.Files.List().
		Fields(createField("id", "name", "webContentLink", "size"), googleapi.Field("nextPageToken")).
		Q("mimeType='audio/mpeg'").
		Pages(ctx, func(fs *drive.FileList) error {
			for _, f := range fs.Files {
				i++
				f1 = &file{
					Name: f.Name,
					Link: f.WebContentLink,
					Tags: []string{},
				}
				filenames[f.Id] = f1
			}
			return nil
		})
	fmt.Println("i: ", i)
	fmt.Println("len:", len(filenames))
	db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(user))
		b := root.Bucket([]byte("files"))
		err := b.ForEach(func(k, v []byte) error {
			id := string(k)
			if f, ok := filenames[id]; ok {
				var tags []string
				err := json.Unmarshal(v, &tags)
				if err != nil {
					return err
				}
				f.Tags = tags
			} else {
				b.Delete(k)
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return filenames
}

func handleFilesApi(c *gin.Context) {
	idb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	id := string(idb)
	fs := getFiles(id)
	if fs == nil {
		c.JSON(204, gin.H{})
	}
	c.JSON(200, fs)
}

func makeSearch(id, name, tag string) map[string]*file {
	fnames := getFiles(id)
	if fnames == nil {
		return nil
	}
	for id, f := range fnames {
		if !strings.Contains(strings.ToLower(f.Name), name) {
			delete(fnames, id)
		}
		if tag != "" {
			var found bool
			for _, t := range f.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				delete(fnames, id)
			}
		}
	}
	return fnames
}

func handleSearchApi(c *gin.Context) {
	name := c.Query("name")
	tag := c.Query("tag")
	fmt.Printf("N==%s,T==%s\n", name, tag)
	idb, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	id := string(idb)
	fnames := makeSearch(id, name, tag)
	if fnames == nil {
		c.JSON(204, gin.H{})
	}
	c.JSON(200, fnames)
}

func handleAddTagApi(c *gin.Context) {
	fmt.Println("Handling tags")
	songb := c.Param("id")
	song := []byte(songb)
	tag := c.Param("tag")
	user, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	fmt.Printf("USER: %s, tag: %s, id: %s\n", string(user), tag, songb)

	err := db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket(user)
		if root == nil {
			fmt.Println("ops")
		}
		b := root.Bucket([]byte("files"))
		if b == nil {
			fmt.Println("ops2")
		}
		v := b.Get(song)
		if v == nil {
			tags := []string{tag}
			v, err := json.Marshal(tags)
			if err != nil {
				fmt.Println("22")
				return err
			}
			err = b.Put(song, v)
			if err != nil {
				fmt.Println("Didnt put")
				return err
			}
			fmt.Println("Tags: ", tags)
			return nil
		}
		var tags []string
		err := json.Unmarshal(v, &tags)
		if err != nil {
			fmt.Println("321")
			return err
		}
		for _, t := range tags {
			if t == tag {
				fmt.Println("nope")
				return nil
			}
		}
		tags = append(tags, tag)
		v, err = json.Marshal(tags)
		if err != nil {
			fmt.Println("ahao")
			return err
		}
		err = b.Put(song, v)
		if err != nil {
			fmt.Println("ewq")
			return err
		}
		fmt.Println("Tags2: ", tags)
		return nil
	})

	if err != nil {
		fmt.Println("SMTH")
		fmt.Println(err.Error())
	}
}

func handleDeleteTagApi(c *gin.Context) {
	fmt.Println("Handling tags")
	songb := c.Param("id")
	song := []byte(songb)
	tag := c.Param("tag")
	user, e := ioutil.ReadAll(c.Request.Body)
	if e != nil {
		fmt.Println("err: ", e.Error())
	}
	fmt.Printf("USER: %s, tag: %s, id: %s\n", string(user), tag, songb)

	err := db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket(user)
		if root == nil {
			fmt.Println("ops")
		}
		b := root.Bucket([]byte("files"))
		if b == nil {
			fmt.Println("ops2")
		}
		v := b.Get(song)
		if v == nil {
			return nil
		}
		var tags []string
		err := json.Unmarshal(v, &tags)
		if err != nil {
			fmt.Println("321")
			return err
		}
		for i, t := range tags {
			if t == tag {
				tags = append(tags[:i], tags[i+1:]...)
			}
		}
		v, err = json.Marshal(tags)
		if err != nil {
			fmt.Println("ahao")
			return err
		}
		err = b.Put(song, v)
		if err != nil {
			fmt.Println("ewq")
			return err
		}
		fmt.Println("Tags2: ", tags)
		return nil
	})

	if err != nil {
		fmt.Println("SMTH")
		fmt.Println(err.Error())
	}
}

func handleFileApi(c *gin.Context) {
	id := c.Param("id")
	idb, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	user := string(idb)
	serv := services[user]
	if serv == nil {
		fmt.Println("FAIL")
		c.JSON(204, gin.H{})
	}
	resp, err := serv.Files.Get(id).Download()
	if err != nil {
		fmt.Println(err.Error())
	}
	io.Copy(c.Writer, resp.Body)
}
