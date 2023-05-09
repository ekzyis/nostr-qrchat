console.log(`Your session keys:\n${nsec} ${npub}`);

for (const el of document.querySelectorAll(".npub")) {
  el.textContent = npub.slice(0, 10) + ".." + npub.slice(-10);
}

document.getElementById("chat-form").addEventListener("submit", (event) => {
  event.preventDefault();
  const form = event.target;
  const msg = form.elements["msg"].value;
  const url = "/chat?" + new URLSearchParams({ msg });
  fetch(url, {
    method: "POST",
  });
  form.reset();
});
