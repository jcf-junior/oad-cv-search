const API_URL = "";

async function request(path, options) {
  const response = await fetch(API_URL + path, options);
  const data = await response.json().catch(() => null);
  if (!response.ok) {
    throw new Error(data?.error || `Erro ${response.status}`);
  }
  return data;
}

function createElement(tag, className, text) {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
}

function setStatus(element, message, isError = false) {
  element.textContent = message;
  element.classList.toggle("error", isError);
}

function excerpt(content) {
  const text = content.replace(/\s+/g, " ").trim();
  return text.length > 200 ? text.slice(0, 200) + "…" : text;
}

function renderResult({ candidate, scoreLabel, barPercent }) {
  const details = createElement("details", "result");
  const summary = createElement("summary");

  const title = createElement("div", "result-title");
  title.append(
    createElement("span", "", `${candidate.name}`),
    createElement("span", "score", scoreLabel),
  );
  summary.append(title);

  if (barPercent !== undefined) {
    const bar = createElement("div", "bar");
    const fill = createElement("div", "bar-fill");
    fill.style.width = `${barPercent}%`;
    bar.append(fill);
    summary.append(bar);
  }

  summary.append(createElement("p", "excerpt", excerpt(candidate.content)));
  details.append(summary, createElement("pre", "", candidate.content));
  return details;
}

function renderResults(container, items) {
  container.replaceChildren(...items.map(renderResult));
}

function semanticItems(results) {
  const best = Math.max(...results.map((r) => r.similarity));
  return results.map((r) => ({
    candidate: r,
    scoreLabel: `similaridade ${r.similarity.toFixed(3)}`,
    barPercent: best > 0 ? (r.similarity / best) * 100 : 0,
  }));
}

function keywordItems(results) {
  return results.map((r, index) => ({
    candidate: r,
    scoreLabel: `#${index + 1} · rank ${r.rank.toFixed(4)}`,
  }));
}

function setupSearch({ prefix, path, toItems }) {
  const form = document.getElementById(`${prefix}-form`);
  const input = document.getElementById(`${prefix}-input`);
  const status = document.getElementById(`${prefix}-status`);
  const results = document.getElementById(`${prefix}-results`);
  const button = form.querySelector("button");

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    button.disabled = true;
    results.replaceChildren();
    setStatus(status, "A pesquisar...");

    try {
      const data = await request(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query: input.value, limit: 5 }),
      });

      if (data.length === 0) {
        setStatus(status, "Sem resultados.");
        return;
      }

      setStatus(status, `${data.length} resultado(s).`);
      renderResults(results, toItems(data));
    } catch (error) {
      setStatus(status, error.message, true);
    } finally {
      button.disabled = false;
    }
  });
}

function setupUpload() {
  const form = document.getElementById("upload-form");
  const fileInput = document.getElementById("upload-file");
  const status = document.getElementById("upload-status");
  const resultBox = document.getElementById("upload-result");
  const meta = document.getElementById("upload-meta");
  const content = document.getElementById("upload-content");
  const button = form.querySelector("button");

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    button.disabled = true;
    resultBox.hidden = true;
    setStatus(status, "A processar: extração de texto e geração do embedding...");

    const body = new FormData();
    body.append("file", fileInput.files[0]);

    try {
      const candidate = await request("/candidates", { method: "POST", body });
      meta.textContent = `${candidate.file_name} · id ${candidate.id}`;
      content.textContent = candidate.content;
      resultBox.hidden = false;
      setStatus(status, "CV guardado.");
    } catch (error) {
      setStatus(status, error.message, true);
    } finally {
      button.disabled = false;
    }
  });
}

function setupTabs() {
  const tabs = document.querySelectorAll(".tab");
  const panels = document.querySelectorAll(".panel");

  tabs.forEach((tab) => {
    tab.addEventListener("click", () => {
      tabs.forEach((t) => t.classList.toggle("active", t === tab));
      panels.forEach((panel) => {
        panel.hidden = panel.id !== tab.dataset.tab;
      });
    });
  });
}

setupTabs();
setupSearch({ prefix: "semantic", path: "/search", toItems: semanticItems });
setupSearch({ prefix: "keyword", path: "/search/keyword", toItems: keywordItems });
setupUpload();
