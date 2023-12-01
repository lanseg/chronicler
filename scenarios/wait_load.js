const imgs = [...document.querySelectorAll("img")];

// Wait for all images to load
const imagesReady = imgs.map((img) => img.complete).reduce((a, b) => a & b, true);

return Boolean(imagesReady);
