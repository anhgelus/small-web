import htmx from "htmx.org";

htmx.config.historyRestoreAsHxRequest = false;

function setupAnchors() {
    document.querySelectorAll("a").forEach(e => {
        // stuff related to external links are already handled in the backend
        if (!e.href.startsWith(window.location.origin) && /https?:\/\//.test(e.href)) return
        if (e.hasAttribute("hx-trigger")) return;
        e.setAttribute("hx-get", e.href)
        e.setAttribute("hx-trigger", "click")
        e.setAttribute("hx-target", "#content")
        e.setAttribute("hx-swap", "outerHTML show:body:top")
        htmx.process(e)
    });
}

// updating history and window title
document.addEventListener("htmx:afterSettle", e => {
    if (e.detail.xhr === undefined) return
    const title = e.detail.xhr.getResponseHeader("Updated-Title")
    if (title?.length !== 0) document.title = title
    const quote = e.detail.xhr.getResponseHeader("Updated-Quote")
    if (quote?.length !== 0)
        document.querySelector("#quote")!.innerHTML = "«&thinsp;" + quote + "&thinsp;»"
    setupAnchors()
})

document.body.addEventListener('htmx:beforeSwap', function(e) {
    if(e.detail.xhr.status !== 404) return
    e.detail.shouldSwap = true;
    e.detail.isError = false;
})

setupAnchors()
