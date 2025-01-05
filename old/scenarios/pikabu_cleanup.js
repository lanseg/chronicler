const selectors = [".stories-feed", ".overlay", ".sidebar", ".theme-picker", "header", "footer"];

// Remove garbage
selectors.forEach((s) => document.querySelectorAll(s).forEach((e) => e.remove()));
const count = selectors.map((s) => document.querySelectorAll(s).length).reduce((a, b) => a + b, 0);

// Remove size limitations
document.querySelectorAll(".app__inner").forEach((e) => {
    e.style["max-width"] = "100%";
});

return Boolean(count === 0);
