let header = document.querySelector("#header");
let headers = header.querySelectorAll("li");
let csrf = document.querySelector("#csrf").content;
let lastActiveHeader;

let leftBtn = document.querySelector("#pagination #left");
let statusTxt = document.querySelector("#pagination #status");
let rightBtn = document.querySelector("#pagination #right");
let page = 0;
let maxPage = 0;
let currentId = 0;
let currentKind = "";
let currentHeader;
const count = 30;

leftBtn.addEventListener("click", e => {
    if (page > 0) {
        page--;
        updateStatus();
        updateCards(currentId, currentKind);
        if (page == 0) {
            leftBtn.setAttribute("disabled", "");
        }
        if (page < maxPage) {
            rightBtn.removeAttribute("disabled");
        }
    }
})
updateStatus();
rightBtn.addEventListener("click", e => {
    if (page < maxPage) {
        page++;
        updateStatus();
        updateCards(currentId, currentKind);
        if (page == maxPage) {
            rightBtn.setAttribute("disabled", "");
        }
        if (page > 0) {
            leftBtn.removeAttribute("disabled");
        }
    }
})

for (let hdr of headers) {
    hdr.addEventListener("click", e => {
        let id = hdr.dataset.id;
        let kind = hdr.dataset.kind;
        if (lastActiveHeader == hdr) return;
        page = 0;
        currentId = id;
        currentKind = kind;
        currentHeader = e.currentTarget;
        updateCards(id, kind);
        contentCount = queryContentCount();
    });
}

function updateStatus() {
    statusTxt.textContent = `${page+1}/${maxPage+1}`;
}

function queryContentCount() {
    fetch(`/api/v1/content?kind=${currentKind}&accountID=${currentId}`, {
        method: "GET",
        headers: {
            "csrf": csrf,
        },
    })
    .then(resp => resp.json())
    .then(data => {
        maxPage = Math.floor(data.Count / count);
        if (data.Count > 0 && data.Count % count == 0) {
            maxPage--
        }
        if (maxPage > 0) {
            rightBtn.removeAttribute("disabled");
        } else {
            leftBtn.setAttribute("disabled", "");
            rightBtn.setAttribute("disabled", "");
        }
        updateStatus();
    });
}

function updateCards(id, kind) {
    fetch(`/partial-renderer/cards?id=${id}&kind=${kind}&page=${page}&count=${count}`, {
        method: "GET",
        headers: {
            "csrf": csrf,
        },
    })
    .then(resp => resp.text())
    .then(data => {
        currentId = id;
        currentKind = kind;
        document.querySelector("#cards").outerHTML = data;
        lastActiveHeader.classList.remove("active");
        currentHeader.classList.add("active");
        lastActiveHeader = currentHeader;
        initCarousel();
    });
}

headers[0].click();
lastActiveHeader = headers[0];