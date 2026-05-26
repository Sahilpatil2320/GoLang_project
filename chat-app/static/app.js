const socket = new WebSocket("ws://" + window.location.host + "/ws/" + room);

socket.onmessage = function(event) {
    const msg = JSON.parse(event.data);
    const chatBox = document.getElementById("chat-box");

    const align = msg.username === username ? "right" : "left";
    const messageDiv = document.createElement("div");
    messageDiv.className = `message ${align}`;
    messageDiv.innerHTML = `<strong>${msg.username}:</strong> ${msg.message}`;
    chatBox.appendChild(messageDiv);

    chatBox.scrollTop = chatBox.scrollHeight;
};

function sendMessage() {
    const messageInput = document.getElementById("message");
    const message = messageInput.value.trim();
    if (message === "") return;

    socket.send(JSON.stringify({ message: message }));
    messageInput.value = "";
    messageInput.focus();
}

// Press Enter to send message
document.getElementById("message").addEventListener("keypress", function (e) {
    if (e.key === "Enter") {
        sendMessage();
    }
});
