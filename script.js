const menuButton = document.querySelector(".menu-button");
const navLinks = document.querySelector(".nav-links");

if (menuButton && navLinks) {
    menuButton.addEventListener("click", () => {
        navLinks.classList.toggle("active");
    });
}

document.querySelectorAll(".nav-links a").forEach(link => {
    link.addEventListener("click", () => {
        navLinks.classList.remove("active");
    });
});


// ================= AI CHURCH ASSISTANT =================

document.addEventListener("DOMContentLoaded", () => {

    const aiForm = document.getElementById("ai-form");
    const questionInput = document.getElementById("ai-question");
    const chatBox = document.getElementById("ai-chat-box");

    if (!aiForm) {
        return;
    }

    aiForm.addEventListener("submit", async (e) => {

        e.preventDefault();

        const question = questionInput.value.trim();

        if (!question) {
            return;
        }

        // Show visitor's question
        const userMessage = document.createElement("div");

        userMessage.className = "ai-message ai-user";

        userMessage.innerHTML = `
            <div class="ai-message-name">
                You
            </div>

            <div class="ai-message-text">
                ${question}
            </div>
        `;

        chatBox.appendChild(userMessage);

        // Clear input
        questionInput.value = "";

        // Scroll to bottom
        chatBox.scrollTop = chatBox.scrollHeight;

        // Show thinking message
        const thinkingMessage = document.createElement("div");

        thinkingMessage.className = "ai-message ai-bot";

        thinkingMessage.innerHTML = `
            <div class="ai-message-name">
                Church Assistant
            </div>

            <div class="ai-message-text">
                Assistant is thinking...
            </div>
        `;

        chatBox.appendChild(thinkingMessage);

        chatBox.scrollTop = chatBox.scrollHeight;

        try {

            const formData = new URLSearchParams();

            formData.append("question", question);

            const response = await fetch("/ai", {
                method: "POST",
                body: formData,
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            });

            if (!response.ok) {
                throw new Error("Server returned an error.");
            }

            const answer = await response.text();

            // Replace thinking message with AI answer
            thinkingMessage.querySelector(".ai-message-text").textContent = answer;

        } catch (error) {

            console.error("AI error:", error);

            thinkingMessage.querySelector(".ai-message-text").textContent =
                "Sorry, I am unable to answer right now. Please contact the church directly.";
        }

        chatBox.scrollTop = chatBox.scrollHeight;
    });

});