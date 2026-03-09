import("./bootstrap").catch((err) => {
  const root = document.getElementById("root");
  if (root) {
    root.innerHTML = `<div style="padding:40px;color:red;font-family:monospace"><h1>Failed to load app</h1><pre>${err.message}\n${err.stack}</pre></div>`;
  }
});
