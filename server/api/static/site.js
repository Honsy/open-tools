(function () {
  var recentKey = "ot-recent";
  var customKey = "ot-custom";

  function load(key) {
    try {
      var value = JSON.parse(localStorage.getItem(key) || "[]");
      return Array.isArray(value) ? value : [];
    } catch (err) {
      return [];
    }
  }

  function save(key, list) {
    localStorage.setItem(key, JSON.stringify(list));
  }

  function tile(link) {
    var a = document.createElement("a");
    a.className = "link tile";
    a.href = link.slug ? "/site/" + link.slug : link.url;
    var mark = document.createElement("span");
    mark.className = "mark";
    mark.textContent = (link.name || "站").slice(0, 1);
    var meta = document.createElement("span");
    meta.className = "meta";
    var name = document.createElement("span");
    name.className = "name";
    name.textContent = link.name || "";
    meta.appendChild(name);
    a.appendChild(mark);
    a.appendChild(meta);
    return a;
  }

  function fill(id, list, empty) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = "";
    if (!list.length) {
      el.textContent = empty;
      return;
    }
    list.forEach(function (link) {
      el.appendChild(tile(link));
    });
  }

  function show(box, id) {
    box.querySelectorAll("[data-tab]").forEach(function (btn) {
      btn.classList.toggle("on", btn.getAttribute("data-tab") === id);
    });
    box.querySelectorAll("[data-panel]").forEach(function (panel) {
      panel.hidden = panel.getAttribute("data-panel") !== id;
    });
  }

  document.querySelectorAll("[data-tabs]").forEach(function (box) {
    box.querySelectorAll("[data-tab]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        show(box, btn.getAttribute("data-tab"));
      });
    });
  });

  var hash = "";
  try {
    hash = decodeURIComponent(location.hash.slice(1));
  } catch (err) {
    hash = "";
  }
  if (hash.indexOf("tab-") === 0) {
    var tab = document.getElementById(hash);
    if (tab) tab.click();
    var scroll = tab && tab.getAttribute("data-scroll");
    var section = scroll && document.getElementById(scroll);
    if (section) section.scrollIntoView();
  } else if (hash) {
    var node = document.getElementById(hash);
    if (node) node.scrollIntoView();
  }
  if (hash && history.replaceState) {
    history.replaceState(null, "", location.pathname + location.search);
  }

  var boards = document.querySelectorAll("#sec-boards [data-tab]");
  if (boards.length > 1) {
    var index = 0;
    setInterval(function () {
      index = (index + 1) % boards.length;
      boards[index].click();
    }, 5000);
  }

  document.addEventListener("click", function (event) {
    var star = event.target.closest("[data-star]");
    if (!star) return;
    event.preventDefault();
    event.stopPropagation();
    var link = JSON.parse(star.getAttribute("data-star"));
    var custom = load(customKey);
    var exists = custom.some(function (item) { return item.id === link.id; });
    custom = exists
      ? custom.filter(function (item) { return item.id !== link.id; })
      : custom.concat([link]).slice(-16);
    save(customKey, custom);
    document.querySelectorAll("[data-star]").forEach(function (btn) {
      var item = JSON.parse(btn.getAttribute("data-star"));
      var on = custom.some(function (row) { return row.id === item.id; });
      btn.classList.toggle("on", on);
      btn.setAttribute("aria-label", on ? "取消收藏" : "收藏");
    });
    fill("desk-custom", custom, "点星星，把网站留在这里。");
  });

  document.addEventListener("click", function (event) {
    var linkEl = event.target.closest("a.link");
    if (!linkEl || event.target.closest("[data-star]")) return;
    var raw = linkEl.getAttribute("data-link");
    if (!raw) return;
    var link = JSON.parse(raw);
    var recent = load(recentKey).filter(function (item) { return item.id !== link.id; });
    recent.unshift(link);
    recent = recent.slice(0, 12);
    save(recentKey, recent);
    fill("desk-recent", recent, "还没有打开过网站。");
  });

  var custom = load(customKey);
  fill("desk-recent", load(recentKey), "还没有打开过网站。");
  fill("desk-custom", custom, "点星星，把网站留在这里。");
  document.querySelectorAll("[data-star]").forEach(function (btn) {
    var item = JSON.parse(btn.getAttribute("data-star"));
    if (custom.some(function (row) { return row.id === item.id; })) btn.classList.add("on");
  });

  var form = document.querySelector("form.search");
  if (!form) return;
  var engines = {
    baidu: "https://www.baidu.com/s?wd=",
    so: "https://www.so.com/s?q=",
    sogou: "https://www.sogou.com/web?query=",
    bing: "https://www.bing.com/search?q=",
    google: "https://www.google.com/search?q=",
    bili: "https://search.bilibili.com/all?keyword=",
    zhihu: "https://www.zhihu.com/search?type=content&q=",
  };
  form.addEventListener("submit", function (event) {
    var picked = form.querySelector("select").value;
    var text = form.querySelector("input[name=q]").value.trim();
    if (!engines[picked] || !text) return;
    event.preventDefault();
    window.open(engines[picked] + encodeURIComponent(text), "_blank", "noopener");
  });
})();
