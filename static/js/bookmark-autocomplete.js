var url = document.getElementById('url');

var pasted = false;
var timeout = null;

suggest = function(remoteRes) {
	if (isValidURL(remoteRes)) {
		var toField = document.getElementById('name');
		suggest_title(remoteRes, toField);
	}
}

url.addEventListener('input', function(e) {
	var remoteRes = this.value;
	clearTimeout(timeout);
	timeout = setTimeout(function() {
		if (!pasted) suggest(remoteRes); else pasted = false; }, 1000);
});

url.addEventListener('paste', function(e) {
	pasted = true;
	var remoteRes = e.clipboardData.getData('text');
	clearTimeout(timeout);
	suggest(remoteRes);
});


function suggest_title(remoteRes, e) {
	console.log("Checking ", remoteRes);

	var requestURL = "/short-title?webpage=" + encodeURIComponent(remoteRes);
	console.log("RequestURL:", requestURL);

	var xhr = new XMLHttpRequest();
	xhr.onreadystatechange = function() { if (this.readyState == 4 &&
	this.status == 200) {
	    var title = xhr.responseText;
		e.value = title;
	} }
	xhr.open("GET", requestURL);
	xhr.send();
}

function isValidURL(u) {
	var a = document.createElement('a');
	a.href = u;
	return (a.host && a.host != window.location.host);
}
