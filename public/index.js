console.log(`Your session keys:\n${nsec} ${npub}`);

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
