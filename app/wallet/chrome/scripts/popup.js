var nonce = 0;

window.onload = function () {
    wireEvents();
    // showInfoTab("send");    
    connect();
}

// =============================================================================

function wireEvents() {
    const refresh = document.getElementById("refreshsubmit");
    refresh.addEventListener(
        'click',
        load,
        false
    );

    const from = document.getElementById("from");
    from.addEventListener(
        'change',
        load,
        false
    );

    const to = document.getElementById("to");
    to.addEventListener(
        'change',
        load,
        false
    );

    const send = document.getElementById("sendbutton");
    send.addEventListener(
        'click',
        showInfoTabSend,
        false
    );

    const tran = document.getElementById("tranbutton");
    tran.addEventListener(
        'click',
        showInfoTabTran,
        false
    );

    const memp = document.getElementById("mempbutton");
    memp.addEventListener(
        'click',
        showInfoTabMemp,
        false
    );

    const sendsubmit = document.getElementById("sendsubmit");
    sendsubmit.addEventListener(
        'click',
        submitTran,
        false
    );

    const sendamount = document.getElementById("sendamount");
    sendamount.addEventListener(
        'keyup',
        formatCurrencyKeyup,
        false
    );
    sendamount.addEventListener(
        'blur',
        formatCurrencyBlur,
        false
    );

    const closebuttonconf = document.getElementById("closebuttonconf");
    closebuttonconf.addEventListener(
        'click',
        closeModal,
        false
    );

    const closebuttonmsg = document.getElementById("closebuttonmsg");
    closebuttonmsg.addEventListener(
        'click',
        closeModal,
        false
    );

    const confirmno = document.getElementById("confirmno");
    confirmno.addEventListener(
        'click',
        closeModal,
        false
    );

    const confirmyes = document.getElementById("confirmyes");
    confirmyes.addEventListener(
        'click',
        createTransaction,
        false
    );
}

// =============================================================================

function connect() {
    var socket = new WebSocket('ws://localhost:8080/v1/events');

    socket.addEventListener('open', function (event) {
        const conn = document.getElementById("connected");
        conn.className = "connected";
        conn.innerHTML = "CONNECTED";
        load();
    });

    socket.addEventListener('close', function (event) {
        const conn = document.getElementById("connected");
        conn.className = "notconnected";
        conn.innerHTML = "NOT CONNECTED";
    });

    socket.addEventListener('message', function (event) {
        const conn = document.getElementById("connected");

        if (event.data.includes("MINING: completed")) {
            conn.className = "connected";
            conn.innerHTML = "CONNECTED";
            load();
            return;
        }

        if (event.data.includes("MINING")) {
            conn.className = "mining";
            conn.innerHTML = "MINING...";
            return;
        }
    });

    socket.addEventListener('error', function (event) {
        const conn = document.getElementById("connected");
        conn.className = "notconnected";
        conn.innerHTML = "NOT CONNECTED";
        showMessage("Unable to connect to node.");
    });
}
