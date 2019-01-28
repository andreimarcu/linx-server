// @license magnet:?xt=urn:btih:1f739d935676111cfff4b4693e3816e664797050&dn=gpl-3.0.txt GPL-v3-or-Later

var navlist = document.getElementById("info").getElementsByClassName("info-actions")[0];

init();

function init() {
    var editA = document.createElement('a');

    editA.setAttribute("href", "#");
    editA.addEventListener('click', function(ev) {
        edit(ev);
        return false;
    });
    editA.innerHTML = "edit";

    var separator = document.createTextNode(" | ");
    navlist.insertBefore(editA, navlist.firstChild);
    navlist.insertBefore(separator, navlist.children[1]);

    document.getElementById('save').addEventListener('click', paste);
    document.getElementById('wordwrap').addEventListener('click', wrap);
}

function edit(ev) {
    ev.preventDefault();

    navlist.remove();
    document.getElementById("filename").remove();
    document.getElementById("editform").style.display = "block";

    var normalcontent = document.getElementById("normal-content");
    normalcontent.removeChild(document.getElementById("normal-code"));

    var editordiv = document.getElementById("inplace-editor");
    editordiv.style.display = "block";
    editordiv.addEventListener('keydown', handleTab);
}

function paste(ev) {
    var editordiv = document.getElementById("inplace-editor");
    document.getElementById("newcontent").value = editordiv.value;
    document.forms["reply"].submit();
}

function wrap(ev) {
    if (document.getElementById("wordwrap").checked) {
        document.getElementById("codeb").style.wordWrap = "break-word";
        document.getElementById("codeb").style.whiteSpace = "pre-wrap";
    }

    else {
        document.getElementById("codeb").style.wordWrap = "normal";
        document.getElementById("codeb").style.whiteSpace = "pre";
    }
}

// @license-end
