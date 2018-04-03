const cookie_name = "JAMPY_USER_ID";
var data;

function makeTag(id) {
    return function (tag) {
        return  '<span class="tag-item">' +
                    tag +
                    `<img onclick="delTag('` + tag + `','` + id + `')" class="del-tag" src="static/grey_cross.png" title="Remove Tag">` +
                '</span>';
    }
}

const sum = (x, y) => x + y;

function makeSong(d) {
    let song = d[1];
    let id = d[0];
    return  '<div id="' + id + '" class="song">' +
                `<div class="play-button" onclick="playSong('`+ song.name +`','` + id + `')">` +
                    '<img src="static/grey-sound.jpg">' +
                '</div>' +
                '<div class="song-inner-wrapper">' +
                    '<div class="song-name">' +
                        song.name +
                    '</div>' +
                    '<div class="tag-list">' +
                        song.tags.map(makeTag(id)).reduce(sum, "") +
                        '<img class="add-img" onclick="showInp(this)" src="static/add_plus.png" title="Add New Tag">'+
                        '<input class="add-inp" onchange="addTag(this,\'' + id + '\')" placeholder="Add Tag" style="display: none">'+
                    '</div>' +
                '</div>' +
            '</div>';
}

function showInp(img) {
    let i = $(img);
    i.hide();
    i.parents().children("input").show();
    i.parents().children("input").focus();

}

function addTag(txt, id) {
    console.log("ID:", id);
    $.post("/api/files/" + id + "/" + txt.value, $.cookie(cookie_name));
    txt.value = "";

    makePage();
}

function delTag(tag, id) {
    console.log("del...");
    $.ajax({
        url: "/api/files/" + id + "/" + tag,
        method: 'DELETE',
        data: $.cookie(cookie_name)
    });
    makePage();
}

function playSong(name, i) {
    $("#jquery_jplayer_1").jPlayer("setMedia", {
        title: name,
        mp3: "/file/" + i
    });
}

function makePage() {
    const user_id = $.cookie(cookie_name);
    $.ajax({
        url: "/api/files",
        type: "post",
        data: user_id,
        statusCode: {
            200: function (js) {
                data = js;
                if (Object.keys(data).length === 0) {
                    $("#app").html("No songs, sorry");
                } else {
                    $("#app").html('<div class="song-list">' +
                        Object.entries(data).map(makeSong).reduce(sum) +
                        '</div>')
                }
            },
            204: () => {
                $.cookie(cookie_name, "", {
                    expires: -1
                });
                window.location.reload()
            }
        }
    });
}


function search(inp) {
    var txt = inp.value;

    if(txt === "") {
        makePage();
        return
    }
    console.log(txt, txt[0] === '#', txt.substring(1));
    const user_id = $.cookie(cookie_name);
    $.ajax({
        url: "/api/search?"+ (txt[0] === '#' ? "tag="+txt.substring(1) : "name="+txt),
        type: "post",
        data: user_id,
        statusCode: {
            200: function (js) {
                data = js;
                if (Object.keys(data).length === 0) {
                    $("#app").html("No songs, sorry");
                } else {
                    $("#app").html('<div class="song-list">' +
                        Object.entries(data).map(makeSong).reduce(sum) +
                        '</div>')
                }
            },
            204: () => {
                $.cookie(cookie_name, "", {
                    expires: -1
                });
                window.location.reload()
            }
        }
    });
}

$(document).ready(() => {
    makePage();
    $("#logout").click(() => {
        $.cookie(cookie_name, "", {
            expires: -1
        });
        window.location.reload()
    });
    $("#jquery_jplayer_1").jPlayer({
        swfPath: "jplayer/jquery.jplayer.swf",
        supplied: "mp3",
        wmode: "window",
        useStateClassSkin: true,
        autoBlur: false,
        smoothPlayBar: true,
        keyEnabled: true,
        remainingDuration: true,
        toggleDuration: true,
    });
});
