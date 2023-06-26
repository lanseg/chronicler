export function pad(number, len = 2) {
    return String(number).padStart(2, '0');
}

export function formatDateTime(timestamp) {
    if (!timestamp) {
        return "No date";
    }
    const date = new Date(timestamp * 1000);
    return `${pad(date.getDate())}.${pad(date.getMonth() + 1)}.${date.getFullYear()} ` +
        `${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function createElement(name, attributes) {
    const el = document.createElement(name);
    Object.entries(attributes ?? {}).map(obj => {
        el.setAttribute(obj[0], obj[1]);
    });
    return el;
}