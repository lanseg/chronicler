export function pad(number) {
    return String(number).padStart(2, '0');
}

export function formatDateTime(date) {
    if (!date) {
        return "No date";
    }
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