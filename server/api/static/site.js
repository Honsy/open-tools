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
    a.className = "site-link card card-compact tile bg-base-200 hover:bg-base-300 shadow-none";
    a.href = link.slug ? "/site/" + link.slug : link.url;
    var raw = JSON.stringify({
      id: link.id, name: link.name, url: link.url, desc: link.desc,
      slug: link.slug, categorySlug: link.categorySlug, clicks: link.clicks
    });
    a.setAttribute("data-link", raw);
    var mark = document.createElement("span");
    mark.className = "mark";
    mark.textContent = (link.name || "站").slice(0, 1);
    var meta = document.createElement("span");
    meta.className = "meta";
    var name = document.createElement("span");
    name.className = "name";
    name.textContent = link.name || "";
    var star = document.createElement("button");
    star.type = "button";
    star.className = "star";
    star.textContent = "★";
    star.setAttribute("aria-label", "收藏");
    star.setAttribute("data-star", raw);
    meta.appendChild(name);
    a.appendChild(mark);
    a.appendChild(meta);
    a.appendChild(star);
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
      btn.classList.toggle("tab-active", btn.getAttribute("data-tab") === id);
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

  var catOpen = document.getElementById("cat-open");
  var catPop = document.getElementById("cat-pop");
  function setCat(open) {
    if (!catPop || !catOpen) return;
    catPop.hidden = !open;
    catOpen.setAttribute("aria-expanded", open ? "true" : "false");
  }
  if (catOpen) {
    catOpen.addEventListener("click", function (event) {
      event.stopPropagation();
      setCat(catPop.hidden);
    });
  }
  var hoverMenu = window.matchMedia("(hover: hover) and (pointer: fine)").matches;
  if (!hoverMenu) {
    document.querySelectorAll(".dropdown > button").forEach(function (btn) {
      btn.addEventListener("click", function (event) {
        event.preventDefault();
        event.stopPropagation();
        var drop = btn.parentElement;
        var open = drop.classList.contains("dropdown-open");
        document.querySelectorAll(".dropdown.dropdown-open").forEach(function (node) {
          node.classList.remove("dropdown-open");
          var toggle = node.querySelector("button");
          if (toggle) toggle.setAttribute("aria-expanded", "false");
        });
        if (!open) {
          drop.classList.add("dropdown-open");
          btn.setAttribute("aria-expanded", "true");
        }
      });
    });
  }
  document.addEventListener("click", function (event) {
    if (catPop && !catPop.hidden && !catPop.contains(event.target) && event.target !== catOpen) setCat(false);
    if (!hoverMenu && !event.target.closest(".dropdown")) {
      document.querySelectorAll(".dropdown.dropdown-open").forEach(function (node) {
        node.classList.remove("dropdown-open");
        var toggle = node.querySelector("button");
        if (toggle) toggle.setAttribute("aria-expanded", "false");
      });
    }
  });
  document.addEventListener("keydown", function (event) {
    if (event.key !== "Escape") return;
    setCat(false);
    document.querySelectorAll(".dropdown.dropdown-open").forEach(function (node) { node.classList.remove("dropdown-open"); });
  });

  var simpleKey = "ot-simple";
  function applySimple(on) {
    document.documentElement.classList.toggle("simple", on);
    document.querySelectorAll("[data-simple]").forEach(function (btn) {
      btn.classList.toggle("btn-active", on);
    });
  }
  applySimple(localStorage.getItem(simpleKey) === "1");
  document.querySelectorAll("[data-simple]").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var on = !document.documentElement.classList.contains("simple");
      localStorage.setItem(simpleKey, on ? "1" : "0");
      applySimple(on);
    });
  });

  var signedIn = !!document.body.getAttribute("data-user");
  var state = { stars: [], recent: [] };
  var starEmpty = "点星星，把网站留在这里。";
  var recentEmpty = "还没有打开过网站。";

  function paintStars(list) {
    document.querySelectorAll("[data-star]").forEach(function (btn) {
      var item = JSON.parse(btn.getAttribute("data-star"));
      var on = list.some(function (row) { return row.id === item.id; });
      btn.classList.toggle("on", on);
      btn.setAttribute("aria-label", on ? "取消收藏" : "收藏");
    });
  }

  function applyDesk(stars, recent) {
    state.stars = stars || [];
    state.recent = recent || [];
    fill("desk-recent", state.recent, recentEmpty);
    fill("desk-custom", state.stars, starEmpty);
    paintStars(state.stars);
  }

  function openAuth(mode) {
    var id = mode === "register" ? "register-dialog" : "login-dialog";
    var box = document.getElementById(id);
    if (!box || typeof box.showModal !== "function") {
      location.href = (mode === "register" ? "/register" : "/login") + "?next=" + encodeURIComponent(location.pathname + location.search);
      return;
    }
    ["login-dialog", "register-dialog"].forEach(function (other) {
      var el = document.getElementById(other);
      if (el && el !== box && el.open) el.close();
    });
    var field = box.querySelector("[name=next]");
    if (field) field.value = location.pathname + location.search;
    var error = box.querySelector("[data-auth-error]");
    if (error) {
      error.hidden = true;
      error.textContent = "";
    }
    if (!box.open) box.showModal();
    var input = box.querySelector("[name=username]");
    if (input) input.focus();
  }

  document.querySelectorAll("[data-auth-open]").forEach(function (link) {
    link.addEventListener("click", function (event) {
      event.preventDefault();
      openAuth(link.getAttribute("data-auth-open") || "login");
    });
  });
  document.querySelectorAll("#login-dialog form[action], #register-dialog form[action]").forEach(function (form) {
    form.addEventListener("submit", function (event) {
      event.preventDefault();
      var error = form.parentElement.querySelector("[data-auth-error]");
      var button = form.querySelector("[type=submit]");
      if (button) button.disabled = true;
      var body = new URLSearchParams(new FormData(form));
      body.set("next", location.pathname + location.search);
      fetch(form.action, {
        method: "POST",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "Accept": "application/json",
          "X-Auth-Modal": "1"
        },
        body: body
      }).then(function (res) {
        return res.json().then(function (data) { return { ok: res.ok, data: data }; });
      }).then(function (result) {
        if (button) button.disabled = false;
        if (result.ok) {
          var next = (result.data && result.data.next) || "/";
          if (next === location.pathname + location.search) location.reload();
          else location.href = next;
          return;
        }
        if (error) {
          error.hidden = false;
          error.textContent = (result.data && result.data.error) || "没有登录上";
        }
      }).catch(function () {
        if (button) button.disabled = false;
        if (error) {
          error.hidden = false;
          error.textContent = "没有发出去，再试一次";
        }
      });
    });
  });

  function api(method, path, body) {
    return fetch(path, {
      method: method,
      credentials: "same-origin",
      headers: body ? { "Content-Type": "application/json" } : {},
      body: body ? JSON.stringify(body) : undefined
    }).then(function (res) {
      if (res.status === 401) {
        openAuth("login");
        return null;
      }
      return res.json();
    });
  }

  function sameId(list, id) {
    return list.some(function (row) { return row.id === id; });
  }

  document.addEventListener("click", function (event) {
    var star = event.target.closest("[data-star]");
    if (!star) return;
    event.preventDefault();
    event.stopPropagation();
    var link = JSON.parse(star.getAttribute("data-star"));
    if (signedIn) {
      api("POST", "/api/me/star", { id: link.id }).then(function (res) {
        if (res) applyDesk(res.stars, state.recent);
      });
      return;
    }
    var custom = load(customKey);
    var exists = sameId(custom, link.id);
    custom = exists
      ? custom.filter(function (item) { return item.id !== link.id; })
      : custom.concat([link]).slice(-16);
    save(customKey, custom);
    applyDesk(custom, load(recentKey));
  });

  document.addEventListener("click", function (event) {
    var linkEl = event.target.closest("a.site-link");
    if (!linkEl || event.target.closest("[data-star]")) return;
    var raw = linkEl.getAttribute("data-link");
    if (!raw) return;
    var link = JSON.parse(raw);
    if (signedIn) {
      api("POST", "/api/me/recent", { id: link.id }).then(function (res) {
        if (res) applyDesk(state.stars, res.recent);
      });
      return;
    }
    var recent = load(recentKey).filter(function (item) { return item.id !== link.id; });
    recent.unshift(link);
    recent = recent.slice(0, 12);
    save(recentKey, recent);
    fill("desk-recent", recent, recentEmpty);
  });

  if (signedIn) {
    var localStars = load(customKey);
    var localRecent = load(recentKey);
    api("GET", "/api/me/desk").then(function (desk) {
      if (!desk) return;
      var chain = Promise.resolve();
      localStars.forEach(function (item) {
        if (!sameId(desk.stars, item.id)) {
          chain = chain.then(function () { return api("POST", "/api/me/star", { id: item.id }); });
        }
      });
      localRecent.slice().reverse().forEach(function (item) {
        chain = chain.then(function () { return api("POST", "/api/me/recent", { id: item.id }); });
      });
      return chain.then(function () {
        localStorage.removeItem(customKey);
        localStorage.removeItem(recentKey);
        return api("GET", "/api/me/desk");
      });
    }).then(function (desk) {
      if (desk) applyDesk(desk.stars, desk.recent);
    });
  } else {
    applyDesk(load(customKey), load(recentKey));
  }

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
