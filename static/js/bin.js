var navlist = document.getElementById("info").getElementsByClassName("right")[0];

init();

function init() {

    var editA = document.createElement('a');

    editA.setAttribute("href", "#");
    editA.setAttribute("onclick", "edit();return false;");
    editA.innerHTML = "edit";

    var separator = document.createTextNode(" | ");
    navlist.insertBefore(editA, navlist.firstChild);
    navlist.insertBefore(separator, navlist.children[1]);

}


function edit() {

    navlist.remove();
    document.getElementById("filename").remove();
    document.getElementById("foarm").style.display = "block";

    var normalcontent = document.getElementById("normal-content");
    normalcontent.removeChild(document.getElementById("normal-code"));
    
    var editordiv = document.getElementById("editor");
    editordiv.style.display = "block";
}


function paste() {
    var editordiv = document.getElementById("editor");
    document.getElementById("newcontent").value = editordiv.value;
    document.forms["reply"].submit();

}

function wrap() {
    if (document.getElementById("wordwrap").checked) {
        document.getElementById("codeb").style.wordWrap = "break-word";
        document.getElementById("codeb").style.whiteSpace = "pre-wrap";
    }

    else {
        document.getElementById("codeb").style.wordWrap = "normal";
        document.getElementById("codeb").style.whiteSpace = "pre";
    }
}
