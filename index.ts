import 'htmx.org';

document.querySelectorAll("a").forEach(e => {
    if (e.href.startsWith(window.location.origin)) return;
    e.target = "_blank";
});
