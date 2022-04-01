var nonce = 0;

window.onload = function () {
    wireEvents();
    showInfoTab("send");    
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

    // const sendsubmit = document.getElementById("sendsubmit");
    // sendsubmit.addEventListener(
    //     'click',
    //     submitTran,
    //     false
    // );

    // const sendamount = document.getElementById("sendamount");
    // sendamount.addEventListener(
    //     'keyup',
    //     formatCurrencyKeyup,
    //     false
    // );
    // sendamount.addEventListener(
    //     'blur',
    //     formatCurrencyBlur,
    //     false
    // );

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

    // const confirmyes = document.getElementById("confirmyes");
    // confirmyes.addEventListener(
    //     'click',
    //     createTransaction,
    //     false
    // );
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

// =============================================================================

$.ajaxSetup({
    contentType: "application/json; charset=utf-8",
    beforeSend: function () {
        closeModal();
    }
});

function handleAjaxError(jqXHR, exception) {
    var msg = '';

    switch (jqXHR.status) {
    case 0:
        msg = 'Not connected, verify network.';
    case 404:
        msg = 'Requested page not found. [404]';
    case 500:
        msg = 'Internal Server Error [500].';
    default:
        switch (exception) {
        case "parsererror":
            msg = 'Requested JSON parse failed.';
        case "timeout":
            msg = 'Time out error.';
        case "abort":
            msg = 'Ajax request aborted.';
        default:
            const o = JSON.parse(jqXHR.responseText);
            msg = o.error;
        }
    }

    showMessage(msg);
}

// ==============================================================================

function load() {
    const conn = document.getElementById("connected");
    if (conn.innerHTML != "CONNECTED") {
        showMessage("No connection to node.");
        return;
    }

    nonce = 0;
    document.getElementById("tranbutton").innerHTML = "Trans";

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/genesis/list",
        success: function (response) {
            fromBalance();
            // toBalance();
            // transactions();
            // mempool();
        },
        error: function (jqXHR, exception) {
            showMessage(exception);
        },
    });
}

function fromBalance() {
    const wallet = new ethers.Wallet(document.getElementById("from").value);

    $.ajax({
        type: "get",
        url: "http://localhost:8080/v1/accounts/list/" + wallet.address,
        success: function (resp) {
            const bal = document.getElementById("frombal");
            bal.innerHTML = formatter.format(resp.accounts[0].balance) + " ARD";

            document.getElementById("fromnonce").innerHTML = resp.accounts[0].nonce;
            if (nonce == 0) {
                nonce = Number(resp.accounts[0].nonce);
                nonce += 1
            }
            document.getElementById("nextnonce").innerHTML = nonce;
        },
        error: function (jqXHR, exception) {
            handleAjaxError(jqXHR, exception);
        },
    });
}


// =============================================================================

function showConfirmation() {
    const modal = document.getElementById("confirmationmodal");
    modal.style.display = "block";

    document.getElementById("yesnomessage").innerHTML = "";
}

function showMessage(msg) {
    const modal = document.getElementById("messagemodal");
    modal.style.display = "block";

    document.getElementById("msg").innerHTML = msg;
}

function closeModal() {
    const confirmationmodal = document.getElementById("confirmationmodal");
    confirmationmodal.style.display = "none";
    const messagemodal = document.getElementById("messagemodal");
    messagemodal.style.display = "none";
    document.getElementById("msg").innerHTML = "";
}

function onConfirm() {

}

// =============================================================================

var formatter = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  
    // These options are needed to round to whole numbers if that's what you want.
    // minimumFractionDigits: 0, // (this suffices for whole numbers, but will print 2500.10 as $2,500.1)
    maximumFractionDigits: 0, // (causes 2500.99 to be printed as $2,501)
});

// =============================================================================

function showInfoTabSend() {
    showInfoTab("send");
}

function showInfoTabTran() {
    showInfoTab("tran");
}

function showInfoTabMemp() {
    showInfoTab("memp");
}

function showInfoTab(which) {
    const sendBox = document.querySelector("div.sendbox");
    const tranBox = document.querySelector("div.tranbox");
    const mempBox = document.querySelector("div.mempbox");

    const sendBut = document.getElementById("sendbutton");
    const tranBut = document.getElementById("tranbutton");
    const mempBut = document.getElementById("mempbutton");

    switch (which) {
    case "send":
        sendBox.style.display = "block";
        tranBox.style.display = "none";
        mempBox.style.display = "none";
        sendBut.style.backgroundColor = "#faf9f5";
        tranBut.style.backgroundColor = "#d9d8d4";
        mempBut.style.backgroundColor = "#d9d8d4";
        break;
    case "tran":
        tranBox.style.display = "block";
        sendBox.style.display = "none";
        mempBox.style.display = "none";
        tranBut.style.backgroundColor = "#faf9f5";
        sendBut.style.backgroundColor = "#d9d8d4";
        mempBut.style.backgroundColor = "#d9d8d4";
        break;
    case "memp":
        mempBox.style.display = "block";
        tranBox.style.display = "none";
        sendBox.style.display = "none";
        mempBut.style.backgroundColor = "#faf9f5";
        tranBut.style.backgroundColor = "#d9d8d4";
        sendBut.style.backgroundColor = "#d9d8d4";
        break;
    }
}

// =============================================================================

function formatCurrencyKeyup() {
    formatCurrency($(this));
}

function formatCurrencyBlur() {
    formatCurrency($(this));
}

function formatNumber(n) {
  // format number 1000000 to 1,234,567
  return n.replace(/\D/g, "").replace(/\B(?=(\d{3})+(?!\d))/g, ",")
}

function formatCurrency(input) {
    // appends $ to value, validates decimal side
    // and puts cursor back in right position.
  
    // get input value
    var input_val = input.val();
    
    // don't validate empty input
    if (input_val === "") { return; }
    
    // original length
    var original_len = input_val.length;

    // initial caret position 
    var caret_pos = input.prop("selectionStart");

    if (input_val.indexOf(".") == 0) { return; }
    
    // no decimal entered
    // add commas to number
    // remove all non-digits
    input_val = formatNumber(input_val);
    input_val = "$" + input_val;
  
    // send updated string to input
    input.val(input_val);

    // put caret back in the right position
    var updated_len = input_val.length;
    caret_pos = updated_len - original_len + caret_pos;
    input[0].setSelectionRange(caret_pos, caret_pos);
}