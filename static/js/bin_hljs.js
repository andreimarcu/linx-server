// @license magnet:?xt=urn:btih:1f739d935676111cfff4b4693e3816e664797050&dn=gpl-3.0.txt GPL-v3-or-Later

hljs.tabReplace = '    ';
hljs.initHighlightingOnLoad();

var codeb = document.getElementById("codeb");
var lines = codeb.innerHTML.split("\n");
codeb.innerHTML = "";
for (var i = 0; i < lines.length; i++) {
	var div = document.createElement("div");
	div.innerHTML = lines[i] + "\n";
	codeb.appendChild(div);
};


var ncode = document.getElementById("normal-code");
ncode.className = "linenumbers";
// @license-end
