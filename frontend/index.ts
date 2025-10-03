import htmx from "htmx.org";

function setupAnchors() {
    document.querySelectorAll("a").forEach(e => {
        if (!e.href.startsWith(window.location.origin)) {
            e.target = "_blank";
            return
        }
        if (e.hasAttribute("hx-trigger")) return;
        e.setAttribute("hx-get", e.href)
        e.setAttribute("hx-trigger", "click")
        e.setAttribute("hx-target", "#content")
        e.setAttribute("hx-swap", "outerHTML show:top")
        htmx.process(e)
    });
}

// updating history and window title
document.addEventListener("htmx:afterSettle", e => {
    const title = e.detail.xhr.getResponseHeader("Updated-Title")
    if (title?.length != 0) document.title = title
    window.history.pushState({}, "", e.detail.pathInfo.finalRequestPath)
    setupAnchors()
})

setupAnchors()
