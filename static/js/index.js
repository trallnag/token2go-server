const reloadButton = document.getElementById("button-reload");
const copyButton = document.getElementById("button-copy");
const deleteButton = document.getElementById("button-delete");

const autoCopySwitch = document.getElementById("switch-auto-copy");
const autoDeleteSwitch = document.getElementById("switch-auto-delete");

const tokenInput = document.getElementById("token-input");
const fingerInput = document.getElementById("finger-input");

// State for usage within this script.
var token = null;
var finger = null;

// Indiciates last time token has been updated / retrieved from backend. Used
// by autoDelete to check if to proceed with deletion after timeout.
let lastUpdate = null;

// Seconds to wait for token deletion to take place in case autoDelete is on.
const autoDeleteDelay = 10

const toast = {
  error: "#cb5f59",
  info: "#58abc2",
  success: "#73b479",
  warning: "#f9a951",
};

function snack(color, msg) {
  Toastify({
    text: msg,
    duration: 4_000,
    close: true,
    gravity: "top",
    style: {
      "background": color,
      "border-radius": "var(--border-radius)",
      "box-shadow": "var(--color-shadow) 5px 5px 6px 1px",
      "color": "white",
    },
  }).showToast();
}

function setPreference(key, value, element, snackOnSuccess) {
  if (value === true) {
    element.setAttribute("aria-checked", "true");
    localStorage.setItem(key, "true");
    console.info(`Enabled ${key}.`);
    if (snackOnSuccess) {
      snack(toast.success, "Saved preference");
    }
    return true
  } else {
    element.setAttribute("aria-checked", "false");
    localStorage.setItem(key, "false");
    console.info(`Disabled ${key}.`);
    if (snackOnSuccess) {
      snack(toast.success, "Saved preference");
    }
    return false
  }
}

if (localStorage.getItem("autoCopy") === "false") {
  autoCopy = setPreference("autoCopy", false, autoCopySwitch, false);
} else {
  autoCopy = setPreference("autoCopy", true, autoCopySwitch, false);
}

if (localStorage.getItem("autoDelete") === "false") {
  autoDelete = setPreference("autoDelete", false, autoDeleteSwitch, false);
} else {
  autoDelete = setPreference("autoDelete", true, autoDeleteSwitch, false);
}

if (localStorage.getItem("finger")) {
  finger = localStorage.getItem("finger")
  fingerInput.value = finger
} else {
  fingerInput.value = "..."
}

async function updateToken() {
  try {
    const response = await fetch("token");
    if (!response.ok) {
      throw new Error("Failed to retrieve response from /token endpoint.");
    }

    const contentType = response.headers.get("Content-Type");
    if (!contentType || !contentType.includes("application/json")) {
      throw new Error("Response content from /token endpoint not JSON.");
    }

    const data = await response.json();
    if (!(data.fingerprint && data.secret)) {
      throw new Error("Response from /token endpoint is missing fields.");
    }

    finger = data.fingerprint;
    token = data.secret;

    localStorage.setItem("finger", finger)

    fingerInput.value = data.fingerprint;
    tokenInput.value = data.secret;

    tokenInput.select();

    if (autoCopy && token) {
      navigator.clipboard.writeText(token);
      snack(toast.success, "Copied token to clipboard");
    }

    if (autoDelete) {
      const localLastUpdate = Date.now().toString()
      lastUpdate = localLastUpdate
      setTimeout(() => {
        if (autoDelete && token && (lastUpdate === localLastUpdate)) {
          token = null;
          tokenInput.value = "...";
          console.debug("Deleted token.")
          snack(toast.success, "Deleted token")
        }
      }, autoDeleteDelay * 1000)
    }
  } catch (error) {
    fingerInput.value = "❌";
    tokenInput.value = "❌";
    console.error(error);
    snack(toast.error, error.message.slice(0, -1));
  }
}

window.addEventListener("DOMContentLoaded", updateToken);

reloadButton.addEventListener("click", updateToken);

copyButton.addEventListener("click", () => {
  if (token) {
    navigator.clipboard.writeText(token);
    snack(toast.success, "Copied token to clipboard");
  } else {
    snack(toast.error, "Token unavailable");
  }
});

autoCopySwitch.addEventListener("click", () => {
  if (autoCopySwitch.getAttribute("aria-checked") === "true") {
    setPreference("autoCopy", false, autoCopySwitch, true);
  } else {
    setPreference("autoCopy", true, autoCopySwitch, true);
  }
});

autoDeleteSwitch.addEventListener("click", () => {
  if (autoDeleteSwitch.getAttribute("aria-checked") === "true") {
    setPreference("autoDelete", false, autoDeleteSwitch, true);
  } else {
    setPreference("autoDelete", true, autoDeleteSwitch, true);
  }
});

deleteButton.addEventListener("click", () => {
  if (token) {
    token = null;
    finger = null;
    tokenInput.value = "...";
    snack(toast.success, "Deleted token");
  } else {
    snack(toast.info, "Nothing to delete");
  }
});
