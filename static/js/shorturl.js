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

                copy(resp.shortUrl);
            } else {
                e.target.setAttribute('aria-label', resp.error)
            }
        }
    };
    xhr.send();
});

function copy(someText) {
    var clipboard = new Clipboard('#shorturl', {
        text: function () {
            return someText;
        }
    });

    clipboard.on('success', function (e) {
        e.trigger.setAttribute('aria-label', 'Successfully copied')
    });

    clipboard.on('error', function (e) {
        e.trigger.setAttribute('aria-label', 'Your browser does not support coping to clipboard')
    });
}
