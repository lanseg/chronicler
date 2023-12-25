"use strict";

class Bus {
    listeners = new Map();

    subscribe(name, callback) {
        if (!this.listeners[name]) {
            this.listeners[name] = [];
        }
        this.listeners[name].push(callback);
    }

    publish(name, anEvent) {
        (this.listeners[name] || []).forEach((l) => l(anEvent));
    }
}

export function getBus() {
    return new Bus();
}
