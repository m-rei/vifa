var accountName = document.querySelector("#account-name");
var btnAddAccount = document.querySelector("#btn-add-account");
var btnDelAccount = document.querySelector("#btn-del-account");
var accountSelection = document.querySelector("#account-selection");
var lastAccountSelection = accountSelection.value;
var channelId = document.querySelector("#channel-id");
var btnAddChannel = document.querySelector("#btn-add");
var channelTable = document.querySelector("#channelTable");
var lastAddedAccount = "";
var lastAddedChannel = "";
var csrf = document.querySelector("#csrf").content;

disable(btnAddAccount);
if (!lastAccountSelection) {
    if (kind === "youtube") {
        disable(btnOpmlFile);
    }
    disable(btnAddChannel);
    disable(btnDelAccount);
} else {
    fillTable();
}
accountName.addEventListener("keyup", e => {
    if (e.keyCode == 13) {
        btnAddAccount.click();        
    }
});
btnAddAccount.addEventListener("click", e => {
    let accName = accountName.value;
    if (!accName || accName == lastAddedAccount) {
        return
    }
    disable(accountName);
    fetch("/api/v1/account", {
        method: "POST",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "accountName": accName,
            "kind": kind,
        }),
    })
    .then(resp => resp.json())
    .then(data => {
        lastAccountSelection = data.ID.toString();
        lastAddedAccount = accountName.value;
        disable(btnAddAccount);
        enable(accountName);
        enable(btnDelAccount);
        accountName.value = "";
        accountName.classList.remove("error");
        renderAccountSelection(data.ID);
        if (channelId.value && channelId.value != lastAddedChannel) {
            enable(btnAddAccount);
        }
    })
    .catch(err => {
        enable(accountName);
        accountName.classList.add("error");
    });
});
btnDelAccount.addEventListener("click", e => {
    if (!lastAccountSelection) {
        return;
    }
    disable(btnDelAccount);
    fetch("/api/v1/account", {
        method: "DELETE",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "accountID": lastAccountSelection,
            "kind": kind,
        }),
    })
    .then(resp => resp.json())
    .then(data => {
        lastAddedAccount = "";
        renderAccountSelection(data.LastID);
        if (data.LastID == -1) {
            lastAccountSelection = "";
            if (kind === "youtube") {
                disable(btnOpmlFile);
            }
            disable(btnAddChannel);
        } else {
            enable(btnDelAccount);
            lastAccountSelection = data.LastID.toString();
        }
    })
    .catch(err => {
        enable(btnDelAccount);
    });
});
accountName.addEventListener("input", e => {
    accountName.classList.remove("error");
    if (accountName.value && accountName.value != lastAddedAccount) {
        enable(btnAddAccount);
    } else {
        disable(btnAddAccount);
    }
});
function accountSelectionChangeEvent(e) {
    if (e.target.value != lastAccountSelection) {
        lastAccountSelection = e.target.value;
        lastAddedChannel = "";
        fillTable();
    }
}
accountSelection.addEventListener("change", accountSelectionChangeEvent);
channelId.addEventListener("input", e => {
    channelId.classList.remove("error");
    if (lastAccountSelection) {
        if (channelId.value && channelId.value != lastAddedChannel) {
            enable(btnAddChannel);
        } else {
            disable(btnAddChannel);
        }
    }
});
channelId.addEventListener("keyup", e => {
    if (e.keyCode == 13) {
        btnAddChannel.click();        
    }
});
btnAddChannel.addEventListener("click", (e) => {
    if (!lastAccountSelection) {
        return;
    }
    lastAddedChannel = channelId.value;
    disable(channelId);
    disable(btnAddChannel);
    fetch("/api/v1/channel", {
        method: "POST",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "ChannelID": channelId.value,
            "AccountID": lastAccountSelection,
            "kind": kind,
        }),
    })
    .then(resp => resp)
    .then(resp => {
        enable(channelId);
        enable(btnAddChannel);
        if (resp.ok) {
            channelId.value = ""
            channelId.classList.remove("error");
            fillTable();
        } else {
            channelId.classList.add("error");
        }
    })
    .catch(err => {
        channelId.classList.add("error");
        enable(channelId);
        enable(btnAddChannel);
    })
});

function fillTable() {
    fetch("/partial-renderer/settings-channel-table", {
        method: "POST",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "accountID": lastAccountSelection,
            "kind": kind,
        }),
    })
    .then(resp => resp.text())
    .then(data => {
        document.querySelector("#channelTable").outerHTML = data;
        channelTable = document.querySelector("#channelTable");
    });
}
function deleteChannelFromAccount(id) {
    fetch("/api/v1/channel", {
        method: "DELETE",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "channelID": id,
            "accountID": lastAccountSelection,
            "kind": kind,
        }),
    })
    .then(resp => resp)
    .then(data => {
        lastAddedChannel = "";;
        fillTable();
    });
}
function renderAccountSelection(id) {
    fetch("/partial-renderer/settings-account-selection", {
        method: "POST",
        headers: {
            "csrf": csrf,
        },
        body: JSON.stringify({
            "ID": id,
            "kind": kind,
        }),
    })
    .then(resp => resp.text())
    .then(data => {
        document.querySelector(".select-wrapper").outerHTML = data;
        accountSelection = document.querySelector("#account-selection");
        accountSelection.addEventListener("change", accountSelectionChangeEvent);
        if (lastAccountSelection && kind === "youtube") {
            enable(btnOpmlFile);
        }
        fillTable()
    })
}
function enable(e) {
    e.removeAttribute("disabled");
}
function disable(e) {
    e.setAttribute("disabled", "");
}