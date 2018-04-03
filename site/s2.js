const cookie_name = "JAMPY_USER_ID";
var data;
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
    window.location.reload()
}

function delTag(tag, id) {
    console.log("del...");
    $.ajax({
        url: "/api/files/" + id + "/" + tag,
        method: 'DELETE',
        data: $.cookie(cookie_name)
    });
    window.location.reload()}

function playSong(name, i) {
    $("#track").html(name);
    $("#pause").show();
    $("#play").hide();
    var w = document.getElementById("audio");
    w.src = "/file/"+i+"/x.mp3";
    w.play()
    w.volume = $("#vol").val()/100;
}

function search(inp) {
    var txt = inp.value;

    if(txt === "") {
        window.location.assign('/');
        return
    }
    console.log(txt, txt[0] === '#', txt.substring(1));
    // window.location.assign("example.com");
    window.location.assign("/search?"+ (txt[0] === '#' ? "tag="+txt.substring(1) : "name="+txt));
}

$(document).ready(() => {
    // makePage();
    $("#logout").click(() => {
        $.cookie(cookie_name, "", {
            expires: -1
        });
        window.location.reload()
    });
    var song = document.getElementById("audio");
    song.addEventListener('timeupdate',function (){
        var curtime = song.currentTime;
        var s = parseInt(curtime % 60);
        var m = parseInt((curtime / 60) % 60);
        s = (s >= 10) ? s : "0" + s;
        m = (m >= 10) ? m : "0" + m;
        $("#time").html(m + ':' + s );
    });
    song.addEventListener("ended", function () {
        $("#pause").hide();
        $("#play").show();
        $("#time").html("00:00");
    });
    $('#vol').change(function () {
        document.getElementById('audio').volume = $(this).val()/100;
    });
    var vol_val;
    $("#vol-img").click(function () {
        var vol = $("#vol");
        if (vol.val() == 0) {
            vol.val(vol_val);
            vol_val = 0;
            document.getElementById('audio').volume = vol.val()/100;
            $(this).attr("src", "./static/sound-w.png");
        } else {
            vol_val = vol.val();
            vol.val(0);
            document.getElementById('audio').volume = 0;
            $(this).attr("src", "./static/no-sound-w.png");
        }
    });
});
