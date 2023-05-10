const validSession = !!nsec && !!npub && !!npubRecipient;

const errorEl = document.getElementById("error");

function initSession() {
  console.log(`Session id: ${sessionId}`);
  console.log(`Your session keys:\n${nsec} ${npub}`);
  console.log(`You are writing to:\n${npubRecipient}`);

  const { data: sk } = NostrTools.nip19.decode(nsec);
  const { data: pk1 } = NostrTools.nip19.decode(npub);
  const { data: pk2 } = NostrTools.nip19.decode(npubRecipient);

  const formatKey = (key) => key.slice(0, 10) + ".." + key.slice(-10);

  for (const el of document.querySelectorAll(".npub")) {
    el.textContent = formatKey(npub);
  }
  for (const el of document.querySelectorAll(".npubRecipient")) {
    el.textContent = formatKey(npubRecipient);
  }

  const history = document.getElementById("history");

  const relays = [
    "wss://relay.damus.io/",
    // "wss://nostr-pub.wellorder.net/",
    "wss://relay.snort.social/",
    "wss://relay.nostr.band/",
    "wss://nostr.wine/",
    "wss://nostr.mutinywallet.com/",
    "wss://nostr.bitcoiner.social/",
    // "wss://rsslay.fiatjaf.com/",
    "wss://nostr.rocks/",
  ];
  const pool = new NostrTools.SimplePool();
  const sub = pool.sub(relays, [
    {
      kinds: [4],
      "#p": [pk1],
    },
  ]);

  async function decryptDM(content) {
    return NostrTools.nip04.decrypt(sk, pk2, content);
  }

  function addMessageToHistory(msg, pubkey) {
    const recv = pubkey === pk2;
    const bubble = document.createElement("div");
    bubble.classList.add("chat-bubble");
    bubble.classList.add(recv ? "recv" : "send");
    const text = document.createElement("div");
    text.classList.add("chat-text");
    text.textContent = msg;
    bubble.appendChild(text);
    history.appendChild(bubble);
  }

  sub.on("event", async (event) => {
    const msg = await decryptDM(event.content);
    console.log(`Received msg: content=${msg} id=${event.id}`);
    addMessageToHistory(msg, event.pubkey);
  });

  async function createEncryptedDM(msg) {
    const ciphertext = await NostrTools.nip04.encrypt(sk, pk2, msg);

    let event = {
      kind: 4,
      pubkey: pk1,
      tags: [["p", pk2]],
      content: ciphertext,
      created_at: Math.floor(Date.now() / 1000),
    };
    event.id = NostrTools.getEventHash(event);
    event.sig = NostrTools.signEvent(event, sk);

    let ok = NostrTools.validateEvent(event);
    if (!ok) {
      throw new Error("event validation failed");
    }
    let sigOk = NostrTools.verifySignature(event);
    if (!sigOk) {
      throw new Error("signature verification failed");
    }
    return event;
  }

  async function sendEncryptedDM(msg) {
    const event = await createEncryptedDM(msg);
    console.log(`Sending msg: content=${msg} id=${event.id}`);
    addMessageToHistory(msg, pk1);
    const pubs = pool.publish(relays, event);
    pubs.on("ok", () => {
      // this may be called multiple times, once for every relay that accepts the event
      console.log(`Sending msg: content=${msg} id=${event.id} - OK`);
    });
  }

  document
    .getElementById("chat-form")
    .addEventListener("submit", async (event) => {
      event.preventDefault();

      const form = event.target;
      const msg = form.elements["msg"].value;
      if (!msg) return;
      await sendEncryptedDM(msg)
        .then(() => {
          errorEl.textContent = "";
        })
        .catch((err) => {
          errorEl.textContent = err;
        });

      form.reset();
    });

  setTimeout(() => {
    // notify recipient about new session
    let msg = `new QRchat session`;
    if (sessionId) msg += ` (id=${sessionId})`;
    createEncryptedDM(msg).then((event) => {
      pool.publish(relays, event);
    });
    const prompt = "Hi. Who's there?";
    addMessageToHistory(prompt, pk2);
  }, 3000);
}

if (validSession) {
  initSession();
} else {
  for (const el of document.querySelectorAll(".session")) {
    el.style.display = "none";
  }
  errorEl.textContent = "session not found";
}
