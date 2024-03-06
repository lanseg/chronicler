// At first, loading all comments
const showComments = [...document.querySelectorAll(".comments__more-button:not(.hidden)")].filter(
    (e) => e.checkVisibility(),
);
if (showComments.length > 0) {
    showComments.forEach((e) => e.click());
    return false;
}
const loadMore = [...document.querySelectorAll(".comment__more")].filter((e) =>
    e.checkVisibility(),
);
if (loadMore.length > 0) {
    loadMore.forEach((e) => e.click());
    return false;
}

// Expanding comment branch if any
const expandBranch = [...document.querySelectorAll(".comment-toggle-children__label")].filter((e) =>
    e.checkVisibility(),
);
if (expandBranch.length > 0) {
    expandBranch.forEach((e) => e.click());
    return false;
}

// Then expanding all comments
const expandComments = [
    ...document.querySelectorAll(".comment-toggle-children.comment-toggle-children_collapse"),
].filter((e) => e.checkVisibility());
if (expandComments.length > 0) {
    expandComments.forEach((e) => e.click);
    return false;
}

// Loading lazy images after all
const imgs = [...document.querySelectorAll("img")];
imgs.filter((img) => img.src === "").forEach(
    (img) => img.getAttribute("data-large-image") ?? img.getAttribute("data-src"),
);
return Boolean(imgs.map((img) => img.complete).reduce((a, b) => a & b, true));
