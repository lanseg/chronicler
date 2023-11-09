const showComments = document.querySelectorAll(".comments__more-button:not(.hidden)");
const expandComments = document.querySelectorAll(
    ".comment-toggle-children.comment-toggle-children_collapse",
);
const loadMore = document.querySelectorAll(".comment__more");
const imgs = [...document.querySelectorAll("img")];

// Expand posts and comments
showComments.forEach((e) => e.click());
loadMore.forEach((e) => e.click());
expandComments.forEach((e) => e.click());

// Load all lazy images
imgs.filter((img) => img.src === "").forEach(
    (img) => img.getAttribute("data-large-image") ?? img.getAttribute("data-src"),
);

// Wait for all images to load
const imagesReady = imgs.map((img) => img.complete).reduce((a, b) => a & b, true);

return Boolean(expandComments.length + loadMore.length + showComments.length === 0 && imagesReady);
