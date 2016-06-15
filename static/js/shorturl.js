document.getElementById('shorturl').addEventListener('click', function (e) {
    e.preventDefault();

    if (e.target.href !== "") return;

    xhr = new XMLHttpRequest();
    xhr.open("GET", e.target.dataset.url, true);
    xhr.setRequestHeader('Accept', 'application/json');
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            var resp = JSON.parse(xhr.responseText);

            if (xhr.status === 200 && resp.error == null) {
                e.target.innerText = resp.shortUrl;
                e.target.href = resp.shortUrl;
                e.target.setAttribute('aria-label', 'Click to copy into clipboard')
            } else {
                e.target.setAttribute('aria-label', resp.error)
            }
        }
    };
    xhr.send();
});

var clipboard = new Clipboard("#shorturl", {
    text: function (trigger) {
        if (trigger.href == null) return;

        return trigger.href;
    }
});

clipboard.on('success', function (e) {
    e.trigger.setAttribute('aria-label', 'Successfully copied')
});

clipboard.on('error', function (e) {
    e.trigger.setAttribute('aria-label', 'Your browser does not support coping to clipboard')
});
