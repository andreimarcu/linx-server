Dropzone.options.dropzone = {
	init: function() {
	},
	addedfile: function(file) {
		var upload = document.createElement("div");
		upload.className = "upload";

		var left = document.createElement("span");
		left.innerHTML = file.name;
		file.leftElement = left;
		upload.appendChild(left);

		var right = document.createElement("div");
		right.className = "right";
		var rightleft = document.createElement("span");
		rightleft.className = "cancel";
		rightleft.innerHTML = "Cancel";
		rightleft.onclick = function(ev) {
			this.removeFile(file);
		}.bind(this);

		var rightright = document.createElement("span");
		right.appendChild(rightleft);
		file.rightLeftElement = rightleft;
		right.appendChild(rightright);
		file.rightRightElement = rightright;
		file.rightElement = right;
		upload.appendChild(right);
		file.uploadElement = upload;
		document.getElementById("uploads").appendChild(upload);
	},
	uploadprogress: function(file, p, bytesSent) {
		p = parseInt(p);
		file.rightRightElement.innerHTML = p + "%";
		file.uploadElement.setAttribute("style", 'background-image: -webkit-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -moz-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -ms-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -o-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%)');
	},
	sending: function(file, xhr, formData) {
		formData.append("randomize", document.getElementById("randomize").checked);
		formData.append("expires", document.getElementById("expires").selectedOptions[0].value);
	},
	success: function(file, resp) {
		file.rightLeftElement.innerHTML = "";
		file.leftElement.innerHTML = '<a target="_blank" href="' + resp.url + '">' + resp.url + '</a>';
		file.rightRightElement.innerHTML = "Delete";
		file.rightRightElement.className = "cancel";
		file.rightRightElement.onclick = function(ev) {
		    xhr = new XMLHttpRequest();
			xhr.open("DELETE", resp.url, true);
			xhr.setRequestHeader("X-Delete-Key", resp.delete_key);
			xhr.onreadystatechange = function(file) {
				if (xhr.status === 404) {
					file.leftElement.innerHTML = 'Deleted <a target="_blank" href="' + resp.url + '">' + resp.url + '</a>';
					file.leftElement.className = "deleted";
					file.rightRightElement.onclick = null;
					file.rightRightElement.innerHTML = "";					
				}
			}.bind(this, file);
			xhr.send();
		}.bind(this);
	},
	error: function(file, errMsg, xhrO) {
		file.rightLeftElement.onclick = null;
		file.rightLeftElement.innerHTML = "";
		file.rightRightElement.innerHTML = "";
		if (file.status === "canceled") {
			file.leftElement.innerHTML = "Canceled " + file.name;			
		}
		else {
			file.leftElement.innerHTML = "Could not upload " + file.name;
		}
		file.leftElement.className = "error";
	},

    maxFilesize: 4096,
	previewsContainer: "#uploads",
	parallelUploads: 5,
	headers: {"Accept": "application/json"},
	dictDefaultMessage: "Click or Drop file(s)",
	dictFallbackMessage: ""
};
