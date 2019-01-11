// @license magnet:?xt=urn:btih:1f739d935676111cfff4b4693e3816e664797050&dn=gpl-3.0.txt GPL-v3-or-Later

Dropzone.options.dropzone = {
	init: function() {
		var dzone = document.getElementById("dzone");
		dzone.style.display = "block";
	},
	addedfile: function(file) {
		var upload = document.createElement("div");
		upload.className = "upload";

		var fileLabel = document.createElement("span");
		fileLabel.innerHTML = file.name;
		file.fileLabel = fileLabel;
		upload.appendChild(fileLabel);

		var fileActions = document.createElement("div");
		fileActions.className = "right";
		file.fileActions = fileActions;
		upload.appendChild(fileActions);

		var cancelAction = document.createElement("span");
		cancelAction.className = "cancel";
		cancelAction.innerHTML = "Cancel";
		cancelAction.addEventListener('click', function(ev) {
		    this.removeFile(file);
		}.bind(this));
		file.cancelActionElement = cancelAction;
		fileActions.appendChild(cancelAction);

		var progress = document.createElement("span");
		file.progressElement = progress;
		fileActions.appendChild(progress);

		file.uploadElement = upload;

		document.getElementById("uploads").appendChild(upload);
	},
	uploadprogress: function(file, p, bytesSent) {
		p = parseInt(p);
		file.progressElement.innerHTML = p + "%";
		file.uploadElement.setAttribute("style", 'background-image: -webkit-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -moz-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -ms-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: -o-linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%); background-image: linear-gradient(left, #F2F4F7 ' + p + '%, #E2E2E2 ' + p + '%)');
	},
	sending: function(file, xhr, formData) {
		formData.append("randomize", document.getElementById("randomize").checked);
		formData.append("expires", document.getElementById("expires").value);
	},
	success: function(file, resp) {
        file.fileActions.removeChild(file.progressElement);

        var fileLabelLink = document.createElement("a");
        fileLabelLink.href = resp.url;
        fileLabelLink.target = "_blank";
        fileLabelLink.innerHTML = resp.url;
        file.fileLabel.innerHTML = "";
        file.fileLabelLink = fileLabelLink;
        file.fileLabel.appendChild(fileLabelLink);

        var deleteAction = document.createElement("span");
        deleteAction.innerHTML = "Delete";
        deleteAction.className = "cancel";
		deleteAction.addEventListener('click', function(ev) {
		    xhr = new XMLHttpRequest();
			xhr.open("DELETE", resp.url, true);
			xhr.setRequestHeader("Linx-Delete-Key", resp.delete_key);
			xhr.onreadystatechange = function(file) {
				if (xhr.readyState == 4 && xhr.status === 200) {
				    var text = document.createTextNode("Deleted ");
				    file.fileLabel.insertBefore(text, file.fileLabelLink);
					file.fileLabel.className = "deleted";
					file.fileActions.removeChild(file.cancelActionElement);
				}
			}.bind(this, file);
			xhr.send();
		});
		file.fileActions.removeChild(file.cancelActionElement);
		file.cancelActionElement = deleteAction;
		file.fileActions.appendChild(deleteAction);
	},
	error: function(file, resp, xhrO) {
        file.fileActions.removeChild(file.cancelActionElement);
        file.fileActions.removeChild(file.progressElement);

		if (file.status === "canceled") {
			file.fileLabel.innerHTML = file.name + ": Canceled ";			
		}
		else {
			if (resp.error) {
				file.fileLabel.innerHTML = file.name + ": " + resp.error;
			}
			else if (resp.includes("<html")) {
				file.fileLabel.innerHTML = file.name + ": Server Error";
			}
			else {
				file.fileLabel.innerHTML = file.name + ": " + resp;
			}
		}
		file.fileLabel.className = "error";
	},

    maxFilesize: Math.round(parseInt(document.getElementById("dropzone").getAttribute("data-maxsize"), 10) / 1024 / 1024),
	previewsContainer: "#uploads",
	parallelUploads: 5,
	headers: {"Accept": "application/json"},
	dictDefaultMessage: "Click or Drop file(s) or Paste image",
	dictFallbackMessage: ""
};

document.onpaste = function(event) {
	var items = (event.clipboardData || event.originalEvent.clipboardData).items;
	for (index in items) {
		var item = items[index];
		if (item.kind === "file") {
			Dropzone.forElement("#dropzone").addFile(item.getAsFile());
		}
	}
};

// @end-license
