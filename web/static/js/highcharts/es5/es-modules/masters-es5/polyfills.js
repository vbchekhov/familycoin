/**
 * @license Highcharts JS v@product.version@ (@product.date@)
 * @module highcharts/polyfills
 *
 * (c) 2009-2025 Torstein Honsi
 *
 * License: www.highcharts.com/license
 */
'use strict';
if (!Array.prototype.includes) {
    // eslint-disable-next-line no-extend-native
    Array.prototype.includes = function (searchElement, fromIndex) {
        return this.indexOf(searchElement, fromIndex) > -1;
    };
}
if (!Array.prototype.find) {
    // eslint-disable-next-line no-extend-native
    Array.prototype.find = function (predicate, thisArg) {
        for (var i = 0; i < this.length; i++) {
            if (predicate.call(thisArg, this[i], i, this)) {
                return this[i];
            }
        }
    };
}
if (!Object.entries) {
    Object.entries = function (obj) {
        var keys = Object.keys(obj), iEnd = keys.length, entries = [];
        for (var i = 0; i < iEnd; ++i) {
            entries.push([keys[i], obj[keys[i]]]);
        }
        return entries;
    };
}
if (!Object.values) {
    Object.values = function (obj) {
        var keys = Object.keys(obj), iEnd = keys.length, values = [];
        for (var i = 0; i < iEnd; ++i) {
            values.push(obj[keys[i]]);
        }
        return values;
    };
}
var ElementPrototype = window.Element.prototype;
if (typeof ElementPrototype.matches !== 'function') {
    ElementPrototype.matches = function matches(selector) {
        var element = this;
        var elements = element.ownerDocument.querySelectorAll(selector);
        var index = 0;
        while (elements[index] && elements[index] !== element) {
            ++index;
        }
        return Boolean(elements[index]);
    };
}
if (typeof ElementPrototype.closest !== 'function') {
    ElementPrototype.closest = function closest(selector) {
        var element = this;
        while (element && element.nodeType === 1) {
            if (element === null || element === void 0 ? void 0 : element.matches(selector)) {
                return element;
            }
            element = element.parentNode || null;
        }
        return null;
    };
}
(function () {
    if (typeof window === 'undefined' ||
        window.CustomEvent ||
        !window.document ||
        !window.Event) {
        return false;
    }
    function CustomEvent(type, params) {
        params = params || { bubbles: false, cancelable: false, detail: undefined };
        var evt = window.document.createEvent('CustomEvent');
        evt.initCustomEvent(type, params.bubbles, params.cancelable, params.detail);
        return evt;
    }
    CustomEvent.prototype = window.Event.prototype;
    window.CustomEvent = CustomEvent;
})();
