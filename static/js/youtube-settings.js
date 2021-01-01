var kind = "youtube";

var opmlFile = document.querySelector("#opml-file");
var btnOpmlFile = document.querySelector("#btn-opml-file");
var msgWait = document.querySelector("#wait");

opmlFile.addEventListener("change", e => {
    if (!lastAccountSelection) {
        return;
    }
    var fd = new FormData();
    fd.append("opml-file", opmlFile.files[0]);
    fd.append("accountID", lastAccountSelection);
    msgWait.classList.remove("d-none");
    disable(btnOpmlFile);
    disable(accountSelection);
    disable(btnDelAccount);
    disable(channelId);
    disable(accountName);
    disable(btnAddAccount);
    fetch("/opmlupload", {
        method: "POST",
        headers: {
            "csrf": csrf,
        },
        body: fd,
    })
    .then(resp => resp)
    .then(resp => {
        enable(btnOpmlFile);
        enable(accountSelection);
        enable(btnDelAccount);
        enable(channelId);
        enable(accountName);
        enable(btnAddAccount);
        if (resp.ok) {
            console.log(lastAccountSelection);
            fillTable();
        }
        msgWait.classList.add("d-none");
    })
    .catch(err => {
        enable(btnOpmlFile);
        enable(accountSelection);
        enable(btnDelAccount);
        enable(channelId);
        enable(accountName);
        enable(btnAddAccount);
        msgWait.classList.add("d-none");
    });
});
btnOpmlFile.addEventListener("click", e => {
    opmlFile.click();
});