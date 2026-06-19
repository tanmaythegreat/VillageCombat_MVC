export function formatNum(n) {
    if (!n) return '0';
    n = Math.floor(n);
    if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
    if (n >= 1_000)     return (n / 1_000).toFixed(0) + 'K';
    return n.toString();
}

export function fmt(n) {
    return (n != null && !isNaN(n)) ? Number(n).toLocaleString() : '0';
}

export function formatTime(seconds) {
    if (!seconds) return '—';
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) return `${h}h ${m}m`;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
}

export function escapeHTML(str) {
    if (str == null) return '';
    return String(str)
        .replace(/&/g,  '&amp;')
        .replace(/</g,  '&lt;')
        .replace(/>/g,  '&gt;')
        .replace(/"/g,  '&quot;')
        .replace(/'/g,  '&#39;');
}

export function setAffordability(btn, affordable) {
    if (!btn) return;
    btn.disabled = !affordable;
    btn.classList.toggle('btn-unaffordable', !affordable);
}

export function makeImgWithFallback(src, alt, imgClass, fbClass, fbEmoji) {
    const img = document.createElement('img');
    img.className = imgClass;
    img.src       = src;
    img.alt       = alt;
    img.onerror   = function () {
        const fb = document.createElement('div');
        fb.className   = fbClass;
        fb.textContent = fbEmoji;
        this.parentNode?.replaceChild(fb, this);
    };
    return img;
}