/**
 * @file Describe the file
 */
/* globals magic, importScript, passport, importScriptList */
(function () {
  var scripts = document.getElementsByTagName("script");
  var length = scripts.length;
  var src = scripts[length - 1].src;
  var pos = src.indexOf("/js/");
  var scriptPath = src.substr(0, pos) + "/js/";
  window.importScriptList = {};
  window.importScript = function (filename) {
    if (!filename) {
      return;
    }
    if (filename.indexOf("http://") === -1 && filename.indexOf("https://") === -1) {
      if (filename.substr(0, 1) === "/") {
        filename = filename.substr(1);
      }
      filename = scriptPath + filename;
    }
    if (filename in importScriptList) {
      return;
    }
    importScriptList[filename] = true;
    document.write('<script src="' + filename + '" type="text/javascript"><\/' + "script>");
  };
})();

passport._define("login_tangram.js", function () {
  var passport = passport || window.passport || {};
  !(function (r) {
    window.Array &&
      !Array.prototype.includes &&
      (Array.prototype.includes = function (r, n) {
        function t(r, n) {
          return r === n || ("number" == typeof r && "number" == typeof n && isNaN(r) && isNaN(n));
        }
        if (null == this) throw new TypeError('"this" is null or not defined');
        var e = Object(this),
          a = e.length >>> 0;
        if (0 === a) return !1;
        for (var o = 0 | n, i = Math.max(o >= 0 ? o : a - Math.abs(o), 0); a > i;) {
          if (t(e[i], r)) return !0;
          i++;
        }
        return !1;
      });
    var n = ["yiyan_intl"];
    r.enTpls = n;
    var t = function (r) {
      var n = new RegExp("(^|&)" + r + "=([^&]*)(&|$)", "i"),
        t = window.location.search.substr(1).match(n);
      return null != t ? unescape(t[2]) : "";
    };
    r.queryTpl = t("tpl");
  })(passport);
  var T,
    baiduInstance,
    baidu = (T = function (a, b) {
      return baidu.dom ? baidu.dom(a, b) : null;
    });
  baidu.version = "2.0.0.1";
  baidu.guid = "$BAIDU$";
  baidu.key = "tangram_guid";
  window[baidu.guid] = window[baidu.guid] || {};
  baidu.check = baidu.check || function () {};
  baidu.lang = baidu.lang || {};
  baidu.forEach = function (a, e, d) {
    var c, f, b;
    if (typeof e == "function" && a) {
      if (typeof a.length == "number") {
        for (c = 0, f = a.length; c < f; c++) {
          b = a[c] || (a.charAt && a.charAt(c));
          e.call(d || null, b, c, a);
        }
      } else {
        if (typeof a == "number") {
          for (c = 0; c < a; c++) {
            e.call(d || null, c, c, c);
          }
        } else {
          if (typeof a == "object") {
            for (c in a) {
              if (a.hasOwnProperty(c)) {
                e.call(d || null, a[c], c, a);
              }
            }
          }
        }
      }
    }
    return a;
  };
  baidu.type = (function () {
    var b = {},
      a = [, "HTMLElement", "Attribute", "Text", , , , , "Comment", "Document", , "DocumentFragment"],
      d = "Array Boolean Date Error Function Number RegExp String",
      c = b.toString;
    baidu.forEach(d.split(" "), function (e) {
      b["[object " + e + "]"] = e.toLowerCase();
      baidu["is" + e] = function (f) {
        return baidu.type(f) == e.toLowerCase();
      };
    });
    return function (f) {
      var e = typeof f;
      return e != "object"
        ? e
        : f == null
          ? "null"
          : f._type_ || b[c.call(f)] || a[f.nodeType] || (f == f.window ? "Window" : "") || "object";
    };
  })();
  baidu.isDate = function (a) {
    return baidu.type(a) == "date" && a.toString() != "Invalid Date" && !isNaN(a);
  };
  baidu.isElement = function (a) {
    return baidu.type(a) == "HTMLElement";
  };
  baidu.isEnumerable = function (a) {
    return a != null && typeof a == "object" && (typeof a.length == "number" || typeof a[0] != "undefined");
  };
  baidu.isNumber = function (a) {
    return baidu.type(a) == "number" && isFinite(a);
  };
  baidu.isPlainObject = function (c) {
    var a,
      b = Object.prototype.hasOwnProperty;
    if (baidu.type(c) != "object") {
      return false;
    }
    if (c.constructor && !b.call(c, "constructor") && !b.call(c.constructor.prototype, "isPrototypeOf")) {
      return false;
    }
    for (a in c) {
    }
    return a === undefined || b.call(c, a);
  };
  baidu.isObject = function (a) {
    return typeof a === "function" || (typeof a === "object" && a != null);
  };
  baidu.extend = function (c, f) {
    var d,
      m,
      k,
      a,
      b,
      g = 1,
      e = arguments.length,
      l = c || {},
      h,
      j;
    baidu.isBoolean(c) && (g = 2) && (l = f || {});
    !baidu.isObject(l) && (l = {});
    for (; g < e; g++) {
      m = arguments[g];
      if (baidu.isObject(m)) {
        for (k in m) {
          a = l[k];
          b = m[k];
          if (a === b) {
            continue;
          }
          if (baidu.isBoolean(c) && c && b && (baidu.isPlainObject(b) || (h = baidu.isArray(b)))) {
            if (h) {
              h = false;
              j = a && baidu.isArray(a) ? a : [];
            } else {
              j = a && baidu.isPlainObject(a) ? a : {};
            }
            l[k] = baidu.extend(c, j, b);
          } else {
            if (b !== undefined) {
              l[k] = b;
            }
          }
        }
      }
    }
    return l;
  };
  baidu.createChain = function (f, d, c) {
    var b = f == "dom" ? "$DOM" : "$" + f.charAt(0).toUpperCase() + f.substr(1);
    var e = Array.prototype.slice;
    var a = (baidu[f] =
      baidu[f] ||
      d ||
      function (g) {
        return baidu.extend(g, baidu[f].fn);
      });
    a.extend = function (g) {
      var h;
      for (h in g) {
        a[h] = function () {
          var k = arguments[0];
          f == "dom" && baidu.type(k) == "string" && (k = "#" + k);
          var j = a(k);
          var i = j[h].apply(j, e.call(arguments, 1));
          return baidu.type(i) == "$DOM" ? i.get(0) : i;
        };
      }
      return baidu.extend(baidu[f].fn, g);
    };
    baidu[f][b] = baidu[f][b] || c || function () {};
    a.fn = baidu[f][b].prototype;
    return a;
  };
  baidu.overwrite = function (a, d, c) {
    for (var b = d.length - 1; b > -1; b--) {
      a.prototype[d[b]] = c(d[b]);
    }
    return a;
  };
  baidu.object = baidu.object || {};
  baidu.object.isPlain = baidu.isPlainObject;
  baidu.createChain("string", function (a) {
    var b = baidu.type(a),
      d = new String(~"string|number".indexOf(b) ? a : b),
      c = String.prototype;
    baidu.forEach(baidu.string.$String.prototype, function (f, e) {
      c[e] || (d[e] = f);
    });
    return d;
  });
  baidu.string.extend({
    trim: (function () {
      var a = new RegExp("(^[\\s\\t\\xa0\\u3000]+)|([\\u3000\\xa0\\s\\t]+\x24)", "g");
      return function () {
        return this.replace(a, "");
      };
    })(),
  });
  baidu.createChain("array", function (d) {
    var c = baidu.array.$Array.prototype,
      b = Array.prototype,
      a;
    baidu.type(d) != "array" && (d = []);
    for (a in c) {
      b[a] || (d[a] = c[a]);
    }
    return d;
  });
  baidu.overwrite(baidu.array.$Array, "concat slice".split(" "), function (a) {
    return function () {
      return baidu.array(Array.prototype[a].apply(this, arguments));
    };
  });
  baidu.array.extend({
    indexOf: function (b, c) {
      baidu.check(".+(,number)?", "baidu.array.indexOf");
      var a = this.length;
      (c = c | 0) < 0 && (c = Math.max(0, a + c));
      for (; c < a; c++) {
        if (c in this && this[c] === b) {
          return c;
        }
      }
      return -1;
    },
  });
  baidu.createChain(
    "Callbacks",
    function (m) {
      var l = {};
      function j(o) {
        var n = (l[o] = {});
        baidu.forEach(o.split(/\s+/), function (p, q) {
          n[p] = true;
        });
        return n;
      }
      m = typeof m === "string" ? l[m] || j(m) : baidu.extend({}, m);
      var d,
        a,
        e,
        c,
        f,
        h,
        g = [],
        i = !m.once && [],
        b = function (n) {
          d = m.memory && n;
          a = true;
          h = c || 0;
          c = 0;
          f = g.length;
          e = true;
          for (; g && h < f; h++) {
            if (g[h].apply(n[0], n[1]) === false && m.stopOnFalse) {
              d = false;
              break;
            }
          }
          e = false;
          if (g) {
            if (i) {
              if (i.length) {
                b(i.shift());
              }
            } else {
              if (d) {
                g = [];
              } else {
                k.disable();
              }
            }
          }
        },
        k = {
          add: function () {
            if (g) {
              var o = g.length;
              (function n(p) {
                baidu.forEach(p, function (q, r) {
                  if (typeof q === "function" && (!m.unique || !k.has(q))) {
                    g.push(q);
                  } else {
                    if (q && q.length) {
                      n(q);
                    }
                  }
                });
              })(arguments);
              if (e) {
                f = g.length;
              } else {
                if (d) {
                  c = o;
                  b(d);
                }
              }
            }
            return this;
          },
          remove: function () {
            if (g) {
              baidu.forEach(arguments, function (n, p) {
                var o;
                while ((o = baidu.array(g).indexOf(n, o)) > -1) {
                  g.splice(o, 1);
                  if (e) {
                    if (o <= f) {
                      f--;
                    }
                    if (o <= h) {
                      h--;
                    }
                  }
                }
              });
            }
            return this;
          },
          has: function (n) {
            return baidu.array(g).indexOf(n) > -1;
          },
          empty: function () {
            g = [];
            return this;
          },
          disable: function () {
            g = i = d = undefined;
            return this;
          },
          disabled: function () {
            return !g;
          },
          lock: function () {
            i = undefined;
            if (!d) {
              k.disable();
            }
            return this;
          },
          locked: function () {
            return !i;
          },
          fireWith: function (o, n) {
            n = n || [];
            n = [o, n.slice ? n.slice() : n];
            if (g && (!a || i)) {
              if (e) {
                i.push(n);
              } else {
                b(n);
              }
            }
            return this;
          },
          fire: function () {
            k.fireWith(this, arguments);
            return this;
          },
          fired: function () {
            return !!a;
          },
        };
      return k;
    },
    function () {},
  );
  baidu.createChain(
    "Deferred",
    function (c) {
      var f = Array.prototype.slice;
      var b = [
          ["resolve", "done", baidu.Callbacks("once memory"), "resolved"],
          ["reject", "fail", baidu.Callbacks("once memory"), "rejected"],
          ["notify", "progress", baidu.Callbacks("memory")],
        ],
        d = "pending",
        e = {
          state: function () {
            return d;
          },
          always: function () {
            a.done(arguments).fail(arguments);
            return this;
          },
          then: function () {
            var g = arguments;
            return baidu
              .Deferred(function (h) {
                baidu.forEach(b, function (j, k) {
                  var m = j[0],
                    l = g[k];
                  a[j[1]](
                    typeof l === "function"
                      ? function () {
                          var i = l.apply(this, arguments);
                          if (i && typeof i.promise === "function") {
                            i.promise().done(h.resolve).fail(h.reject).progress(h.notify);
                          } else {
                            h[m + "With"](this === a ? h : this, [i]);
                          }
                        }
                      : h[m],
                  );
                });
                g = null;
              })
              .promise();
          },
          promise: function (g) {
            return typeof g === "object" ? baidu.extend(g, e) : e;
          },
        },
        a = {};
      e.pipe = e.then;
      baidu.forEach(b, function (g, h) {
        var k = g[2],
          j = g[3];
        e[g[1]] = k.add;
        if (j) {
          k.add(
            function () {
              d = j;
            },
            b[h ^ 1][2].disable,
            b[2][2].lock,
          );
        }
        a[g[0]] = k.fire;
        a[g[0] + "With"] = k.fireWith;
      });
      e.promise(a);
      if (c) {
        c.call(a, a);
      }
      baidu.extend(baidu, {
        when: function (l) {
          var j = 0,
            n = f.call(arguments),
            g = n.length,
            h = g !== 1 || (l && typeof l.promise === "function") ? g : 0,
            q = h === 1 ? l : baidu.Deferred(),
            k = function (s, t, r) {
              return function (i) {
                t[s] = this;
                r[s] = arguments.length > 1 ? f.call(arguments) : i;
                if (r === p) {
                  q.notifyWith(t, r);
                } else {
                  if (!--h) {
                    q.resolveWith(t, r);
                  }
                }
              };
            },
            p,
            m,
            o;
          if (g > 1) {
            p = new Array(g);
            m = new Array(g);
            o = new Array(g);
            for (; j < g; j++) {
              if (n[j] && typeof n[j].promise === "function") {
                n[j]
                  .promise()
                  .done(k(j, o, n))
                  .fail(q.reject)
                  .progress(k(j, m, p));
              } else {
                --h;
              }
            }
          }
          if (!h) {
            q.resolveWith(o, n);
          }
          return q.promise();
        },
      });
      return a;
    },
    function () {},
  );
  baidu._util_ = baidu._util_ || {};
  baidu.global =
    baidu.global ||
    (function () {
      var a = (baidu._util_.$global = window[baidu.guid]);
      return function (c, d, b) {
        if (typeof d != "undefined") {
          d = b && typeof a[c] != "undefined" ? a[c] : d;
          a[c] = d;
        } else {
          if (c && typeof a[c] == "undefined") {
            a[c] = {};
          }
        }
        return a[c];
      };
    })();
  baidu.browser =
    baidu.browser ||
    (function () {
      var b = navigator.userAgent;
      var a = {
        isStrict: document.compatMode == "CSS1Compat",
        isGecko: /gecko/i.test(b) && !/like gecko/i.test(b),
        isWebkit: /webkit/i.test(b),
      };
      try {
        / (\d + \.\d + ) /.test(external.max_version) && (a.maxthon = +RegExp["\x241"]);
      } catch (c) {}
      switch (true) {
        case /msie(\d+\.\d+)/i.test(b):
          a.ie = document.documentMode || +RegExp["\x241"];
          break;
        case /chrome\/(\d+\.\d+)/i.test(b):
          a.chrome = +RegExp["\x241"];
          break;
        case /(\d+\.\d)?(?:\.\d)?\s+safari\/?(\d+\.\d+)?/i.test(b) && !/chrome/i.test(b):
          a.safari = +(RegExp["\x241"] || RegExp["\x242"]);
          break;
        case /firefox\/(\d+\.\d+)/i.test(b):
          a.firefox = +RegExp["\x241"];
          break;
        case /opera(\/|)(\d+(\.\d+)?)(.+?(version\/(\d+(\.\d+)?)))?/i.test(b):
          a.opera = +(RegExp["\x246"] || RegExp["\x242"]);
          break;
      }
      baidu.extend(baidu, a);
      return a;
    })();
  baidu.id = (function () {
    var b = baidu.global("_maps_id"),
      a = baidu.key;
    baidu.global("_counter", 1, true);
    return function (d, h) {
      var g,
        f = baidu.isString(d),
        c = baidu.isObject(d),
        i = c ? d[a] : f ? d : "";
      if (baidu.isString(h)) {
        switch (h) {
          case "get":
            return c ? i : b[i];
            break;
          case "remove":
          case "delete":
            if ((g = b[i])) {
              if (baidu.browser.ie && baidu.isElement(g)) {
                g.removeAttribute(a);
              } else {
                delete g[a];
              }
              delete b[i];
            }
            return i;
            break;
          case "decontrol":
            !(g = b[i]) && c && (d[a] = i = baidu.id());
            i && delete b[i];
            return i;
            break;
          default:
            if (f) {
              (g = b[i]) && delete b[i];
              g && (b[(g[a] = h)] = g);
            } else {
              if (c) {
                i && delete b[i];
                b[(d[a] = h)] = d;
              }
            }
            return h;
        }
      }
      if (c) {
        !i && (b[(d[a] = i = baidu.id())] = d);
        return i;
      } else {
        if (f) {
          return b[d];
        }
      }
      return "TANGRAM__PSP_" + baidu._util_.$global._counter++;
    };
  })();
  baidu.id.key = "tangram_guid";
  baidu.merge = function (e, c) {
    var d = e.length,
      b = 0;
    if (typeof c.length === "number") {
      for (var a = c.length; b < a; b++) {
        e[d++] = c[b];
      }
    } else {
      while (c[b] !== undefined) {
        e[d++] = c[b++];
      }
    }
    e.length = d;
    return e;
  };
  baidu.array.extend({
    unique: function (e) {
      var b = this.length,
        a = this.slice(0),
        d,
        c;
      if ("function" != typeof e) {
        e = function (g, f) {
          return g === f;
        };
      }
      while (--b > 0) {
        c = a[b];
        d = b;
        while (d--) {
          if (e(c, a[d])) {
            a.splice(b, 1);
            break;
          }
        }
      }
      b = this.length = a.length;
      for (d = 0; d < b; d++) {
        this[d] = a[d];
      }
      return this;
    },
  });
  baidu.query =
    baidu.query ||
    (function () {
      var d = / ^ (\w * )# ([\w\ - \$] + ) $ /,
        a = /^#([\w\-\$]+)$/;
      ((rTag = /^\w+$/), (rClass = /^(\w*)\.([\w\-\$]+)$/), (rComboClass = /^(\.[\w\-\$]+)+$/));
      ((rDivider = /\s*,\s*/), (rSpace = /\s+/g), (slice = Array.prototype.slice));
      function c(h, f) {
        var o,
          n,
          e,
          i,
          g,
          m,
          j,
          l,
          k = [];
        if (d.test(h)) {
          e = RegExp.$2;
          g = RegExp.$1 || "*";
          baidu.forEach(f.getElementsByTagName(g), function (p) {
            p.id == e && k.push(p);
          });
        } else {
          if (rTag.test(h) || h == "*") {
            baidu.merge(k, f.getElementsByTagName(h));
          } else {
            if (rClass.test(h)) {
              j = [];
              g = RegExp.$1;
              m = RegExp.$2;
              o = " " + m + " ";
              if (f.getElementsByClassName) {
                j = f.getElementsByClassName(m);
              } else {
                baidu.forEach(f.getElementsByTagName("*"), function (p) {
                  p.className && (" " + p.className + " ").indexOf(o) > -1 && j.push(p);
                });
              }
              if (g && (g = g.toUpperCase())) {
                baidu.forEach(j, function (p) {
                  p.tagName.toUpperCase() === g && k.push(p);
                });
              } else {
                baidu.merge(k, j);
              }
            } else {
              if (rComboClass.test(h)) {
                l = h.substr(1).split(".");
                baidu.forEach(f.getElementsByTagName("*"), function (p) {
                  if (p.className) {
                    o = " " + p.className + " ";
                    n = true;
                    baidu.forEach(l, function (q) {
                      o.indexOf(" " + q + " ") == -1 && (n = false);
                    });
                    n && k.push(p);
                  }
                });
              }
            }
          }
        }
        return k;
      }
      function b(e, g) {
        var f,
          h = e,
          j = "__tangram__",
          i = [];
        if (!g && a.test(h) && (f = document.getElementById(h.substr(1)))) {
          return [f];
        }
        g = g || document;
        if (g.querySelectorAll) {
          if (g.nodeType == 1 && !g.id) {
            g.id = j;
            f = g.querySelectorAll("#" + j + " " + h);
            g.id = "";
          } else {
            f = g.querySelectorAll(h);
          }
          return f;
        } else {
          if (h.indexOf(" ") == -1) {
            return c(h, g);
          }
          baidu.forEach(c(h.substr(0, h.indexOf(" ")), g), function (k) {
            baidu.merge(i, b(h.substr(h.indexOf(" ") + 1), k));
          });
        }
        return i;
      }
      return function (f, h, g) {
        if (!f || typeof f != "string") {
          return g || [];
        }
        var e = [];
        f = f.replace(rSpace, " ");
        g && baidu.merge(e, g) && (g.length = 0);
        baidu.forEach(f.indexOf(",") > 0 ? f.split(rDivider) : [f], function (i) {
          baidu.merge(e, b(i, h));
        });
        return baidu.merge(g || [], baidu.array(e).unique());
      };
    })();
  baidu
    .createChain(
      "dom",
      function (a, b) {
        var d,
          c = new baidu.dom.$DOM(b);
        if (!a) {
          return c;
        }
        if (a._type_ == "$DOM") {
          return a;
        } else {
          if (a.nodeType || a == a.window) {
            c[0] = a;
            c.length = 1;
            return c;
          } else {
            if (a.length && c.toString.call(a) != "[object String]") {
              return baidu.merge(c, a);
            } else {
              if (typeof a == "string") {
                if (a.charAt(0) == "<" && a.charAt(a.length - 1) == ">" && a.length > 2) {
                  if (baidu.dom.createElements) {
                    baidu.merge(c, baidu.dom.createElements(a));
                  }
                } else {
                  baidu.query(a, b, c);
                }
              } else {
                if (typeof a == "function") {
                  return c.ready ? c.ready(a) : c;
                }
              }
            }
          }
        }
        return c;
      },
      function (a) {
        this.length = 0;
        this._type_ = "$DOM";
        this.context = a || document;
      },
    )
    .extend({
      size: function () {
        return this.length;
      },
      splice: function () {},
      get: function (a) {
        if (typeof a == "number") {
          return a < 0 ? this[this.length + a] : this[a];
        }
        return Array.prototype.slice.call(this, 0);
      },
      toArray: function () {
        return this.get();
      },
    });
  baidu.dom.extend({
    ready: (function () {
      var f = false,
        e = [],
        b;
      if (document.addEventListener) {
        b = function () {
          document.removeEventListener("DOMContentLoaded", b, false);
          c();
        };
      } else {
        if (document.attachEvent) {
          b = function () {
            if (document.readyState === "complete") {
              document.detachEvent("onreadystatechange", b);
              c();
            }
          };
        }
      }
      function c() {
        if (!c.isReady) {
          c.isReady = true;
          for (var h = 0, g = e.length; h < g; h++) {
            e[h]();
          }
        }
      }
      function a() {
        try {
          document.documentElement.doScroll("left");
        } catch (g) {
          setTimeout(a, 1);
          return;
        }
        c();
      }
      function d() {
        if (f) {
          return;
        }
        f = true;
        if (document.readyState === "complete") {
          c.isReady = true;
        } else {
          if (document.addEventListener) {
            document.addEventListener("DOMContentLoaded", b, false);
            window.addEventListener("load", c, false);
          } else {
            if (document.attachEvent) {
              document.attachEvent("onreadystatechange", b);
              window.attachEvent("onload", c);
              var g = false;
              try {
                g = window.frameElement == null;
              } catch (h) {}
              if (document.documentElement.doScroll && g) {
                a();
              }
            }
          }
        }
      }
      d();
      return function (g) {
        if (g) {
          c.isReady ? g() : e.push(g);
        }
        return this;
      };
    })(),
  });
  baidu.dom.ready = baidu.dom.fn.ready;
  baidu.support =
    baidu.support ||
    (function () {
      var e = document.createElement("div"),
        d,
        b,
        c;
      e.innerHTML = '<a href="/a" style="top:1px; float: left; opacity: .55">Tangram</a><input type="checkbox">';
      b = e.getElementsByTagName("A")[0];
      c = e.getElementsByTagName("input")[0];
      c.checked = true;
      d = {
        opacity: b.style.opacity === "0.55",
        cssFloat: !!b.style.cssFloat,
        noCloneChecked: c.cloneNode(true).checked,
        noCloneEvent: true,
      };
      if (!e.addEventListener && e.attachEvent && e.fireEvent) {
        e.attachEvent("onclick", function () {
          d.noCloneEvent = false;
        });
        e.cloneNode(true).fireEvent("onclick");
      }
      baidu(function () {
        var h = document.getElementsByTagName("body")[0],
          f = document.createElement("div"),
          a = document.createElement("div"),
          n = "padding: 0; margin: 0; border: ",
          m = "left: 0; top: 0; width: 0px; height: 0px; ",
          j = m + n + "0; visibility: hidden;",
          o = m + n + "5px solid #000; position: absolute;",
          q,
          r,
          l,
          g,
          p;
        f.style.cssText = "position: static;" + j;
        h.insertBefore(f, h.firstChild);
        f.appendChild(a);
        a.style.cssText = "position: absolute;" + j;
        a.innerHTML =
          '<div style="' +
          o +
          'display: bloack;"><div style="' +
          n +
          '0; display: block; overflow: hidden;"></div></div><table style="' +
          o +
          '" cellpadding="0" cellspacing="0"><tr><td></td></tr></table>';
        q = a.firstChild;
        r = q.firstChild;
        p = q.nextSibling;
        d.hasBorderWidth = r.offsetTop >= 5;
        d.hasTableCellBorderWidth = p.rows[0].cells[0].offsetTop >= 5;
        r.style.position = "fixed";
        r.style.top = "20px";
        d.fixedPosition = r.offsetTop === 20 || r.offsetTop === 15;
        d.deleteExpando = true;
        try {
          delete a.test;
        } catch (i) {
          d.deleteExpando = false;
        }
        l = document.createElement("select");
        g = l.appendChild(document.createElement("option"));
        l.disabled = true;
        d.optDisabled = !g.disabled;
        d.optSelected = g.selected;
        a.setAttribute("className", "t");
        a.innerHTML = "  <link/><table></table><a href='/a'>a</a><input type='checkbox'/>";
        var k = a.getElementsByTagName("input")[0];
        d.checkOn = k.value === "on";
        d.htmlSerialize = !!a.getElementsByTagName("link").length;
        d.leadingWhitespace = a.firstChild.nodeType === 3;
        d.getSetAttribute = a.className !== "t";
        d.pixelMargin = true;
        a.innerHTML = "";
        a.style.cssText =
          "box-sizing:border-box;-moz-box-sizing:border-box;-webkit-box-sizing:border-box;padding:1px;border:1px;display:block;width:4px;margin-top:1%;position:absolute;top:1%;";
        if (window.getComputedStyle) {
          d.pixelMargin = (window.getComputedStyle(a, null) || {}).marginTop !== "1%";
        }
        var k = document.createElement("input");
        k.value = "t";
        k.setAttribute("type", "radio");
        d.radioValue = k.value === "t";
        d.hrefNormalized = b.getAttribute("href") === "/a";
        d.style = /top/.test(b.getAttribute("style"));
        d.enctype = !!document.createElement("form").enctype;
        h.removeChild(f);
        f = a = q = r = p = null;
      });
      return d;
    })();
  baidu
    .createChain(
      "event",
      (function () {
        var a = {};
        return function (c, b) {
          switch (baidu.type(c)) {
            case "object":
              return a.originalEvent === c ? a : (a = new baidu.event.$Event(c));
            case "$Event":
              return c;
            case "string":
              var d = new baidu.event.$Event(c);
              typeof b == "object" && baidu.forEach(d, b);
              return d;
          }
        };
      })(),
      function (d) {
        var i, b, g;
        var c = this;
        this._type_ = "$Event";
        if (typeof d == "object" && d.type) {
          c.originalEvent = i = d;
          baidu.forEach(
            "altKey bubbles button buttons cancelable clientX clientY ctrlKey commandKey currentTarget fromElement metaKey screenX screenY shiftKey toElement type view which triggerData".split(
              " ",
            ),
            function (e) {
              c[e] = i[e];
            },
          );
          c.target = c.srcElement = i.srcElement || ((b = i.target) && (b.nodeType == 3 ? b.parentNode : b));
          c.relatedTarget = i.relatedTarget || ((b = i.fromElement) && (b === c.target ? i.toElement : b));
          c.keyCode = c.which = i.keyCode || i.which;
          if (!c.which && i.button !== undefined) {
            c.which = i.button & 1 ? 1 : i.button & 2 ? 3 : i.button & 4 ? 2 : 0;
          }
          var h = document.documentElement,
            a = document.body;
          c.pageX =
            i.pageX ||
            i.clientX +
              ((h && h.scrollLeft) || (a && a.scrollLeft) || 0) -
              ((h && h.clientLeft) || (a && a.clientLeft) || 0);
          c.pageY =
            i.pageY ||
            i.clientY +
              ((h && h.scrollTop) || (a && a.scrollTop) || 0) -
              ((h && h.clientTop) || (a && a.clientTop) || 0);
          c.data;
        }
        typeof d == "string" && (this.type = d);
        this.timeStamp = new Date().getTime();
      },
    )
    .extend({
      stopPropagation: function () {
        var a = this.originalEvent;
        a && (a.stopPropagation ? a.stopPropagation() : (a.cancelBubble = true));
      },
      preventDefault: function () {
        var a = this.originalEvent;
        a && (a.preventDefault ? a.preventDefault() : (a.returnValue = false));
      },
    });
  baidu.dom.extend({
    each: function (c) {
      baidu.check("function", "baidu.dom.each");
      var b,
        a,
        d = this.length;
      for (b = 0; b < d; b++) {
        a = c.call(this[b], b, this[b], this);
        if (a === false || a == "break") {
          break;
        }
      }
      return this;
    },
  });
  baidu.dom.match = (function () {
    var b = /^[\w\#\-\$\.\*]+$/,
      c = document.createElement("DIV");
    c.id = "__tangram__";
    return function (k, e, i) {
      var f,
        h = baidu.array();
      switch (baidu.type(e)) {
        case "$DOM":
          for (var d = k.length - 1; d > -1; d--) {
            for (var j = e.length - 1; j > -1; j--) {
              k[d] === e[j] && h.push(k[d]);
            }
          }
          break;
        case "function":
          baidu.forEach(k, function (m, l) {
            e.call(m, l) && h.push(m);
          });
          break;
        case "HTMLElement":
          baidu.forEach(k, function (l) {
            l == e && h.push(l);
          });
          break;
        case "string":
          var g = baidu.query(e, i || document);
          baidu.forEach(k, function (o) {
            if ((f = a(o))) {
              var m = f.nodeType == 1 ? baidu.query(e, f) : g;
              for (var l = 0, p = m.length; l < p; l++) {
                if (m[l] === o) {
                  h.push(o);
                  break;
                }
              }
            }
          });
          h = h.unique();
          break;
        default:
          h = baidu.array(k).unique();
          break;
      }
      return h;
    };
    function a(f) {
      var d = [],
        e;
      while ((f = f.parentNode)) {
        f.nodeType && d.push(f);
      }
      for (var e = d.length - 1; e > -1; e--) {
        if (d[e].nodeType == 1 || d[e].nodeType == 9) {
          return d[e];
        }
      }
      return null;
    }
  })();
  baidu.dom.extend({
    is: function (a) {
      return baidu.dom.match(this, a).length > 0;
    },
  });
  baidu.dom.extend({
    triggerHandler: function (c, b) {
      var a = baidu._util_.eventBase;
      baidu.forEach(this, function (d) {
        a.fireHandler(d, c, b);
      });
      return this;
    },
  });
  baidu.dom._g = function (a) {
    return baidu.type(a) === "string" ? document.getElementById(a) : a;
  };
  baidu.dom.extend({
    contains: function (b) {
      b = baidu.dom(b);
      if (this.size() <= 0 || b.size() <= 0) {
        return false;
      }
      var a = this[0];
      b = b[0];
      return a.contains ? a != b && a.contains(b) : !!(a.compareDocumentPosition(b) & 16);
    },
  });
  baidu.dom.contains = function (a, b) {
    var c = baidu.dom._g;
    a = c(a);
    b = c(b);
    return a.contains ? a != b && a.contains(b) : !!(a.compareDocumentPosition(b) & 16);
  };
  baidu.dom.extend({
    closest: function (a, c) {
      var b = baidu.array();
      baidu.forEach(this, function (e) {
        var d = [e];
        while ((e = e.parentNode)) {
          e.nodeType && d.push(e);
        }
        d = baidu.dom.match(d, a, c);
        d.length && b.push(d[0]);
      });
      return baidu.dom(b.unique());
    },
  });
  baidu._util_.eventBase = (function () {
    var a = {};
    var e = {};
    var j = window.addEventListener
      ? function (m, k, l) {
          m.addEventListener(k, l, false);
        }
      : window.attachEvent
        ? function (m, k, l) {
            m.attachEvent("on" + k, l);
          }
        : function () {};
    var i = function (n, k, m) {
      var p = baidu.id(n);
      var o = (e[p] = e[p] || {});
      if (o[k]) {
        return;
      }
      o[k] = 1;
      var l = function (t) {
        var r = Array.prototype.slice.call(arguments, 1);
        r.unshift((t = baidu.event(t)));
        if (!t.currentTarget) {
          t.currentTarget = n;
        }
        for (var s = 0, q = m.length; s < q; s += 2) {
          m[s].apply(this, r);
        }
      };
      j(n, k, l);
    };
    var c = function (p, l, r, n, o) {
      var s = function (v) {
        var u = baidu.dom(v.target);
        if (o && !v.data) {
          v.data = o;
        }
        if (v.triggerData) {
          [].push.apply(arguments, v.triggerData);
        }
        if (!n) {
          return (v.result = r.apply(p, arguments));
        }
        if (u.is(n) || u.is(n + " *")) {
          return (v.result = r.apply(baidu.dom(v.target).closest(n)[0], arguments));
        }
      };
      var m = baidu.id(p);
      var q = a[m] || (a[m] = {});
      var k = q[l] || (q[l] = []);
      k.push(s, r);
      i(p, l, k);
      return s;
    };
    var h = function (q, k, t, n) {
      var l;
      if (!(l = baidu.id(q, "get"))) {
        return;
      }
      var s = a[l] || (a[l] = {});
      var r = { mouseenter: "mouseover", mouseleave: "mouseout", focusin: "focus", focusout: "blur" };
      var m = s[k] || (s[k] = []);
      if (r[k]) {
        s[r[k]] = [];
      }
      for (var o = m.length - 1, p; o >= 0; o--) {
        if (((p = m[o]), p === t)) {
          m.splice(o - 1, 2);
        }
      }
    };
    var d = function (n, m) {
      var l;
      if (!(l = baidu.id(n, "get"))) {
        return;
      }
      var o = a[l] || (a[l] = {});
      var k = function (q) {
        var p = o[q] || (o[q] = []);
        for (var r = p.length - 1, s; r >= 0; r -= 2) {
          ((s = p[r]), h(n, q, s));
        }
      };
      if (m) {
        k(m);
      } else {
        for (var m in o) {
          k(m);
        }
      }
    };
    var b = function (s, n, m) {
      var o;
      if (!(o = baidu.id(s, "get"))) {
        return;
      }
      var u = a[o] || (a[o] = {});
      var p = u[n] || (u[n] = []);
      var k = baidu.event({ type: n });
      var t = [k];
      if (m) {
        ((k.triggerData = m), t.push.apply(t, m));
      }
      for (var r = 0, q = p.length; r < q; r += 2) {
        p[r].apply(this, t);
      }
    };
    var g = function (r) {
      var n;
      if (!(n = baidu.id(r, "get"))) {
        return;
      }
      var s = a[n] || (a[n] = {});
      var o = {},
        k;
      for (var q in s) {
        k = o[q] = [];
        ce = s[q];
        for (var p = 1, m = ce.length; p < m; p += 2) {
          k.push(ce[p]);
        }
      }
      return o;
    };
    var f = function (m) {
      switch (m) {
        case "focusin":
        case "focusout":
          if (!/firefox/i.test(navigator.userAgent)) {
            return false;
          }
          var l = {},
            k = m == "focusin" ? "focus" : "blur";
          l[m] = function (r, p) {
            if (typeof r == "function") {
              ((p = r), (r = null));
            }
            var q = this;
            if (!p) {
              return this.triggerHandler(m, r);
            } else {
              var o = function () {
                q.triggerHandler(m);
              };
              baidu.forEach(this, function (s) {
                baidu("textarea,select,input,button,a", s).on(k, o);
              });
              return (this._on(m, r, p), this);
            }
          };
          return (baidu.dom.extend(l), true);
        case "mouseenter":
        case "mouseleave":
          if (/msie/i.test(navigator.userAgent)) {
            return false;
          }
          var l = {},
            k = m == "mouseenter" ? "mouseover" : "mouseout";
          var n = baidu.dom.contains;
          l[m] = function (r, p) {
            if (arguments.length == 0) {
              return this.trigger(m);
            }
            if (typeof r == "function") {
              ((p = r), (r = null));
            }
            var q = this;
            var o = function (s) {
              related = s.relatedTarget;
              if (!related || (related !== this && !n(this, related))) {
                q.triggerHandler(m);
              }
            };
            baidu.forEach(
              this,
              function (s) {
                this.on(k, o);
              },
              this,
            );
            return (this._on(m, r, p), this);
          };
          return (baidu.dom.extend(l), true);
      }
      return false;
    };
    return {
      add: function (o, m, l, k, n) {
        return c(o, m, l, k, n);
      },
      get: function (k) {
        return g(k);
      },
      remove: function (n, m, l, k) {
        var o;
        if ((o = baidu.id(n, "get")) && l && l["_" + o + "_" + m]) {
          ((l = l["_" + o + "_" + m]), delete l["_" + o + "_" + m]);
        }
        if (typeof l == "function") {
          return h(n, m, l, k);
        } else {
          return d(n, m, k);
        }
      },
      removeAll: function (k) {
        return d(k);
      },
      fireHandler: function (m, l, k) {
        return b(m, l, k);
      },
      method: function (n) {
        if (arguments.length > 1) {
          for (var o = 0, k = arguments.length; o < k; o++) {
            this.method(arguments[o]);
          }
          return this;
        }
        if (!f(n)) {
          var m = {};
          m[n] = function (p, l) {
            if (arguments.length == 0) {
              return this.trigger(n);
            } else {
              if (typeof p == "function") {
                ((l = p), (p = null));
              }
              return this._on(n, p, l);
            }
          };
          baidu.dom.extend(m);
        }
      },
      _getEventsLength: function (o, p) {
        var k = 0,
          n;
        if (o) {
          n = a[baidu.id(o[0] || o, "get")];
          if (p) {
            n[p] && (k = n[p].length);
          } else {
            for (var m in n) {
              k += n[m].length;
            }
          }
        } else {
          for (var m in a) {
            n = a[m];
            for (var l in n) {
              k += n[l].length;
            }
          }
        }
        return k / 2;
      },
    };
  })();
  baidu._util_.eventBase.method(
    "blur",
    "change",
    "click",
    "dblclick",
    "error",
    "focus",
    "focusin",
    "focusout",
    "keydown",
    "keypress",
    "keyup",
    "mousedown",
    "mouseenter",
    "mouseleave",
    "mousemove",
    "mouseout",
    "mouseover",
    "mouseup",
    "resize",
    "scroll",
    "select",
    "submit",
    "load",
    "unload",
    "contextmenu",
  );
  baidu.dom.g = function (a) {
    if (!a) {
      return null;
    }
    if ("string" == typeof a || a instanceof String) {
      return document.getElementById(a);
    } else {
      if (a.nodeName && (a.nodeType == 1 || a.nodeType == 9)) {
        return a;
      }
    }
    return null;
  };
  baidu.dom.extend({
    on: function (d, a, f, e) {
      var c = baidu._util_.eventBase;
      var b = { mouseenter: 1, mouseleave: 1, focusin: 1, focusout: 1 };
      if (typeof a == "object" && a) {
        ((e = f), (f = a), (a = null));
      } else {
        if (typeof f == "function") {
          ((e = f), (f = null));
        } else {
          if (typeof a == "function") {
            ((e = a), (a = f = null));
          }
        }
      }
      if (typeof d == "string") {
        d = d.split(/[ ,]+/);
        this.each(function () {
          baidu.forEach(
            d,
            function (g) {
              if (b[g]) {
                baidu(this)[g](f, e);
              } else {
                c.add(this, g, e, a, f);
              }
            },
            this,
          );
        });
      } else {
        if (typeof d == "object") {
          if (e) {
            e = null;
          }
          baidu.forEach(
            d,
            function (h, g) {
              this.on(g, a, f, h);
            },
            this,
          );
        }
      }
      return this;
    },
    _on: function (a, d, c) {
      var b = baidu._util_.eventBase;
      this.each(function () {
        b.add(this, a, c, undefined, d);
      });
      return this;
    },
  });
  baidu.event.on = baidu.on = function (a, c, b) {
    a = baidu.dom.g(a);
    baidu.dom(a).on(c.replace(/^\s*on/, ""), b);
    return a;
  };
  void (function () {
    var u = location ? location.href : document.location.href,
      G = /^(?:about|app|app\-storage|.+\-extension|file|res|widget):$/,
      H = /^\/\//,
      I = /^([\w\+\.\-]+:)(?:\/\/([^\/?#:]*)(?::(\d+)|)|)/,
      t = /#.*$/,
      r = /\[\]$/,
      p = /^(?:GET|HEAD)$/,
      A = /([?&])_=[^&]*/,
      v = /^(.*?):[ \t]*([^\r\n]*)\r?$/gm,
      F = /^[\],:{}\s]*$/,
      f = /(?:^|:|,)(?:\s*\[)+/g,
      a = /\\(?:["\\\/bfnrt]|u[\da-fA-F]{4})/g,
      g = /"[^"\\\r\n]*"|true|false|null|-?(?:\d\d*\.|)\d+(?:[eE][\-+]?\d+|)/g,
      e = ["*/"] + ["*"],
      d = {},
      s = {},
      y = {},
      j = {},
      E = I.exec(u.toLowerCase()) || [];
    function z(N) {
      var L, M;
      if (!N || baidu.type(N) !== "string") {
        return null;
      }
      try {
        if (window.DOMParser) {
          M = new DOMParser();
          L = M.parseFromString(N, "text/xml");
        } else {
          L = new ActiveXObject("Microsoft.XMLDOM");
          L.async = "false";
          L.loadXML(N);
        }
      } catch (O) {
        L = undefined;
      }
      if (!L || !L.documentElement || L.getElementsByTagName("parsererror").length) {
        throw new Error("Invalid XML: " + N);
      }
      return L;
    }
    function q(L) {
      if (!L || baidu.type(L) !== "string") {
        return null;
      }
      L = baidu.string(L).trim();
      if (window.JSON && window.JSON.parse) {
        return window.JSON.parse(L);
      }
      if (F.test(L.replace(a, "@").replace(g, "]").replace(f, ""))) {
        return new Function("return " + L)();
      }
      throw new Error("Invalid JSON: " + L);
    }
    function w(L) {
      if (L && /\S/.test(L)) {
        (
          window.execScript ||
          function (M) {
            window["eval"].call(window, M);
          }
        )(L);
      }
    }
    function B(L) {
      return function (R, P) {
        if (baidu.type(R) !== "string") {
          P = R;
          R = "*";
        }
        var N = R.toLowerCase().split(/\s+/),
          Q,
          S;
        if (baidu.type(P) === "function") {
          for (var M = 0, O; (O = N[M]); M++) {
            Q = /^\+/.test(O);
            Q && (O = O.substr(1) || "*");
            S = L[O] = L[O] || [];
            S[Q ? "unshift" : "push"](P);
          }
        }
      };
    }
    function l(L, V, R) {
      var Q,
        S,
        P,
        M,
        N = L.contents,
        U = L.dataTypes,
        O = L.responseFields;
      for (S in O) {
        if (S in R) {
          V[O[S]] = R[S];
        }
      }
      while (U[0] === "*") {
        U.shift();
        if (Q === undefined) {
          Q = L.mimeType || V.getResponseHeader("content-type");
        }
      }
      if (Q) {
        for (S in N) {
          if (N[S] && N[S].test(Q)) {
            U.unshift(S);
            break;
          }
        }
      }
      if (U[0] in R) {
        P = U[0];
      } else {
        for (S in R) {
          if (!U[0] || L.converters[S + " " + U[0]]) {
            P = S;
            break;
          }
          if (!M) {
            M = S;
          }
        }
        P = P || M;
      }
      if (P) {
        if (P !== U[0]) {
          U.unshift(P);
        }
        return R[P];
      }
    }
    function D(L, N) {
      var R = L.dataTypes.slice(),
        M = R[0],
        V = {},
        U,
        Q;
      L.dataFilter && (N = L.dataFilter(N, L.dataType));
      if (R[1]) {
        for (var O in L.converters) {
          V[O.toLowerCase()] = L.converters[O];
        }
      }
      for (var O = 0, W; (W = R[++O]);) {
        if (W !== "*") {
          if (M !== "*" && M !== W) {
            U = V[M + " " + W] || V["* " + W];
            if (!U) {
              for (var S in V) {
                Q = S.split(" ");
                if (Q[1] === W) {
                  U = V[M + " " + Q[0]] || V["* " + Q[0]];
                  if (U) {
                    if (U === true) {
                      U = V[S];
                    } else {
                      if (V[S] !== true) {
                        W = Q[0];
                        R.splice(O--, 0, W);
                      }
                    }
                    break;
                  }
                }
              }
            }
            if (U !== true) {
              if (U && L["throws"]) {
                N = U(N);
              } else {
                try {
                  N = U(N);
                } catch (P) {
                  return { state: "parsererror", error: U ? P : "No conversion from " + M + " to " + W };
                }
              }
            }
          }
          M = W;
        }
      }
      return { state: "success", data: N };
    }
    function i(M, W, Q, U, S, O) {
      S = S || W.dataTypes[0];
      O = O || {};
      O[S] = true;
      var V,
        R = M[S],
        L = R ? R.length : 0,
        P = M === d;
      for (var N = 0; N < L && (P || !V); N++) {
        V = R[N](W, Q, U);
        if (typeof V === "string") {
          if (!P || O[V]) {
            V = undefined;
          } else {
            W.dataTypes.unshift(V);
            V = i(M, W, Q, U, V, O);
          }
        }
      }
      if ((P || !V) && !O["*"]) {
        V = i(M, W, Q, U, "*", O);
      }
      return V;
    }
    baidu.createChain(
      "ajax",
      function (P, N) {
        if (baidu.object.isPlain(P)) {
          N = P;
          P = undefined;
        }
        N = N || {};
        var V = baidu.ajax.setup({}, N),
          aj = V.context || V,
          M,
          U,
          Z,
          ai = baidu.Deferred(),
          ad = baidu.Callbacks("once memory"),
          R = V.statusCode || {},
          Q = 0,
          ae = {},
          Y = {},
          S = "canceled",
          ag,
          O,
          ab,
          ah = baidu.extend(new baidu.ajax.$Ajax(P, V), {
            readyState: 0,
            setRequestHeader: function (al, am) {
              if (!Q) {
                var ak = al.toLowerCase();
                al = ae[ak] = ae[ak] || al;
                Y[al] = am;
              }
            },
            getAllResponseHeaders: function () {
              return Q === 2 ? ag : null;
            },
            getResponseHeader: function (al) {
              var ak;
              if (Q === 2) {
                if (!O) {
                  O = {};
                  while ((ak = v.exec(ag))) {
                    O[ak[1].toLowerCase()] = ak[2];
                  }
                }
                ak = O[al.toLowerCase()];
              }
              return ak === undefined ? null : ak;
            },
            overrideMimeType: function (ak) {
              !Q && (V.mimeType = ak);
              return this;
            },
            abort: function (ak) {
              ak = ak || S;
              ab && ab.abort(ak);
              X(0, ak);
              return this;
            },
          });
        var W;
        function X(ap, al, aq, am) {
          var an = al,
            ak,
            au,
            ar,
            ao,
            at;
          if (Q === 2) {
            return;
          }
          Q = 2;
          W && clearTimeout(W);
          ab = undefined;
          ag = am || "";
          ah.readyState = ap > 0 ? 4 : 0;
          aq && (ao = l(V, ah, aq));
          if ((ap >= 200 && ap < 300) || ap === 304) {
            if (V.ifModified) {
              at = ah.getResponseHeader("Last-Modified");
              at && (y[U] = at);
              at = ah.getResponseHeader("Etag");
              at && (j[U] = at);
            }
            if (ap === 304) {
              an = "notmodified";
              ak = true;
            } else {
              ak = D(V, ao);
              an = ak.state;
              au = ak.data;
              ar = ak.error;
              ak = !ar;
            }
          } else {
            ar = an;
            if (!an || ap) {
              an = "error";
              ap < 0 && (ap = 0);
            }
          }
          ah.status = ap;
          ah.statusText = "" + (al || an);
          if (ak) {
            ai.resolveWith(aj, [au, an, ah]);
          } else {
            ai.rejectWith(aj, [ah, an, ar]);
          }
          ah.statusCode(R);
          R = undefined;
          ad.fireWith(aj, [ah, an]);
        }
        ai.promise(ah);
        ah.success = ah.done;
        ah.error = ah.fail;
        ah.complete = ad.add;
        ah.statusCode = function (al) {
          if (al) {
            if (Q < 2) {
              for (var ak in al) {
                R[ak] = [R[ak], al[ak]];
              }
            } else {
              ah.always(al[ah.status]);
            }
          }
          return this;
        };
        V.url = ((P || V.url) + "").replace(t, "").replace(H, E[1] + "//");
        V.dataTypes = baidu
          .string(V.dataType || "*")
          .trim()
          .toLowerCase()
          .split(/\s+/);
        if (V.crossDomain == null) {
          Z = I.exec(V.url.toLowerCase());
          V.crossDomain = !!(
            Z &&
            (Z[1] != E[1] ||
              Z[2] != E[2] ||
              (Z[3] || (Z[1] === "http:" ? 80 : 443)) != (E[3] || (E[1] === "http:" ? 80 : 443)))
          );
        }
        if (V.data && V.processData && baidu.type(V.data) !== "string") {
          V.data = baidu.ajax.param(V.data, V.traditional);
        }
        i(d, V, N, ah);
        if (Q === 2) {
          return "";
        }
        M = V.global;
        V.type = V.type.toUpperCase();
        V.hasContent = !p.test(V.type);
        if (!V.hasContent) {
          if (V.data) {
            V.url += (~V.url.indexOf("?") ? "&" : "?") + V.data;
            delete V.data;
          }
          U = V.url;
          if (V.cache === false) {
            var L = new Date().getTime(),
              af = V.url.replace(A, "$1_=" + L);
            V.url = af + (af === V.url ? (~V.url.indexOf("?") ? "&" : "?") + "_=" + L : "");
          }
        }
        if ((V.data && V.hasContent && V.contentType !== false) || N.contentType) {
          ah.setRequestHeader("Content-Type", V.contentType);
        }
        if (V.ifModified) {
          U = U || V.url;
          y[U] && ah.setRequestHeader("If-Modified-Since", y[U]);
          j[U] && ah.setRequestHeader("If-None-Match", j[U]);
        }
        ah.setRequestHeader(
          "Accept",
          V.dataTypes[0] && V.accepts[V.dataTypes[0]]
            ? V.accepts[V.dataTypes[0]] + (V.dataTypes[0] !== "*" ? ", " + e + "; q=0.01" : "")
            : V.accepts["*"],
        );
        for (var aa in V.headers) {
          ah.setRequestHeader(aa, V.headers[aa]);
        }
        if (V.beforeSend && (V.beforeSend.call(aj, ah, V) === false || Q === 2)) {
          return ah.abort();
        }
        S = "abort";
        for (var aa in { success: 1, error: 1, complete: 1 }) {
          ah[aa](V[aa]);
        }
        ab = i(s, V, N, ah);
        if (!ab) {
          X(-1, "No Transport");
        } else {
          ah.readyState = 1;
          if (V.async && V.timeout > 0) {
            W = setTimeout(function () {
              ah.abort("timeout");
            }, V.timeout);
          }
          try {
            Q = 1;
            ab.send(Y, X);
          } catch (ac) {
            if (Q < 2) {
              X(-1, ac);
            } else {
              throw ac;
            }
          }
        }
        return ah;
      },
      function (M, L) {
        this.url = M;
        this.options = L;
      },
    );
    baidu.ajax.settings = {
      url: u,
      isLocal: G.test(E[1]),
      global: true,
      type: "GET",
      contentType: "application/x-www-form-urlencoded; charset=UTF-8",
      processData: true,
      async: true,
      accepts: {
        xml: "application/xml, text/xml",
        html: "text/html",
        text: "text/plain",
        json: "application/json, text/javascript",
        "*": e,
      },
      contents: { xml: /xml/, html: /html/, json: /json/ },
      responseFields: { xml: "responseXML", text: "responseText" },
      converters: { "* text": window.String, "text html": true, "text json": q, "text xml": z },
      flatOptions: { context: true, url: true },
    };
    function K(N, P) {
      var O = baidu.ajax.settings.flatOptions || {},
        L;
      for (var M in P) {
        if (P[M] !== undefined) {
          (O[M] ? N : L || (L = {}))[M] = P[M];
        }
      }
      L && baidu.extend(true, N, L);
    }
    baidu.ajax.setup = function (M, L) {
      if (L) {
        K(M, baidu.ajax.settings);
      } else {
        L = M;
        M = baidu.ajax.settings;
      }
      K(M, L);
      return M;
    };
    function o(N, L, M) {
      M = baidu.type(M) === "function" ? M() : M || "";
      N.push(encodeURIComponent(L) + "=" + encodeURIComponent(M));
    }
    function b(P, M, O, N) {
      if (baidu.type(O) === "array") {
        baidu.forEach(O, function (R, Q) {
          if (N || r.test(M)) {
            o(P, M, R);
          } else {
            b(P, M + "[" + (typeof R === "object" ? Q : "") + "]", R, N);
          }
        });
      } else {
        if (!N && baidu.type(O) === "object") {
          for (var L in O) {
            b(P, M + "[" + L + "]", O[L], N);
          }
        } else {
          o(P, M, O);
        }
      }
    }
    baidu.ajax.param = function (O, N) {
      var L = [];
      if (baidu.type(O) === "array") {
        baidu.forEach(O, function (P) {
          o(L, P.name, P.value);
        });
      } else {
        for (var M in O) {
          b(L, M, O[M], N);
        }
      }
      return L.join("&").replace(/%20/g, "+");
    };
    baidu.ajax.prefilter = B(d);
    baidu.ajax.transport = B(s);
    var h = [],
      c = /(=)\?(?=&|$)|\?\?/,
      m = new Date().getTime();
    baidu.ajax.setup({
      jsonp: "callback",
      jsonpCallback: function () {
        var L = h.pop() || baidu.id.key + "_" + m++;
        this[L] = true;
        return L;
      },
    });
    baidu.ajax.prefilter("json jsonp", function (L, R, W) {
      var V,
        M,
        U,
        P = L.data,
        N = L.url,
        O = L.jsonp !== false,
        S = O && c.test(N),
        Q =
          O &&
          !S &&
          baidu.type(P) === "string" &&
          !(L.contentType || "").indexOf("application/x-www-form-urlencoded") &&
          c.test(P);
      if (L.dataTypes[0] === "jsonp" || S || Q) {
        V = L.jsonpCallback = baidu.type(L.jsonpCallback) === "function" ? L.jsonpCallback() : L.jsonpCallback;
        M = window[V];
        if (S) {
          L.url = N.replace(c, "$1" + V);
        } else {
          if (Q) {
            L.data = P.replace(c, "$1" + V);
          } else {
            if (O) {
              L.url += (/\?/.test(N) ? "&" : "?") + L.jsonp + "=" + V;
            }
          }
        }
        L.converters["script json"] = function () {
          return U[0];
        };
        L.dataTypes[0] = "json";
        window[V] = function () {
          U = arguments;
        };
        W.always(function () {
          window[V] = M;
          if (L[V]) {
            L.jsonpCallback = R.jsonpCallback;
            h.push(V);
          }
          if (U && baidu.type(M) === "function") {
            M(U[0]);
          }
          U = M = undefined;
        });
        return "script";
      }
    });
    baidu.ajax.setup({
      accepts: { script: "text/javascript, application/javascript, application/ecmascript, application/x-ecmascript" },
      contents: { script: /javascript|ecmascript/ },
      converters: {
        "text script": function (L) {
          w(L);
          return L;
        },
      },
    });
    baidu.ajax.prefilter("script", function (L) {
      L.cache === undefined && (L.cache = false);
      if (L.crossDomain) {
        L.type = "GET";
        L.global = false;
      }
    });
    baidu.ajax.transport("script", function (N) {
      if (N.crossDomain) {
        var L,
          M = document.head || document.getElementsByTagName("head")[0] || document.documentElement;
        return {
          send: function (O, P) {
            L = document.createElement("script");
            L.async = "async";
            N.scriptCharset && (L.charset = N.scriptCharset);
            L.src = N.url;
            L.onload = L.onreadystatechange = function (Q, R) {
              if (R || !L.readyState || /loaded|complete/.test(L.readyState)) {
                L.onload = L.onreadystatechange = null;
                M && L.parentNode && M.removeChild(L);
                L = undefined;
                !R && P(200, "success");
              }
            };
            M.insertBefore(L, M.firstChild);
          },
          abort: function () {
            L && L.onload(0, 1);
          },
        };
      }
    });
    var k,
      C = 0,
      n = window.ActiveXObject
        ? function () {
            for (var L in k) {
              k[L](0, 1);
            }
          }
        : false;
    function x() {
      try {
        return new window.XMLHttpRequest();
      } catch (L) {}
    }
    function J() {
      try {
        return new window.ActiveXObject("Microsoft.XMLHTTP");
      } catch (L) {}
    }
    baidu.ajax.settings.xhr = window.ActiveXObject
      ? function () {
          return (!this.isLocal && x()) || J();
        }
      : x;
    void (function (L) {
      baidu.extend(baidu.support, { ajax: !!L, cors: !!L && "withCredentials" in L });
    })(baidu.ajax.settings.xhr());
    if (baidu.support.ajax) {
      baidu.ajax.transport(function (L) {
        if (!L.crossDomain || baidu.support.cors) {
          var M;
          return {
            send: function (S, N) {
              var P,
                R = L.xhr();
              if (L.username) {
                R.open(L.type, L.url, L.async, L.username, L.password);
              } else {
                R.open(L.type, L.url, L.async);
              }
              if (L.xhrFields) {
                for (var O in L.xhrFields) {
                  R[O] = L.xhrFields[O];
                }
              }
              if (L.mimeType && R.overrideMimeType) {
                R.overrideMimeType(L.mimeType);
              }
              if (!L.crossDomain && !S["X-Requested-With"]) {
                S["X-Requested-With"] = "XMLHttpRequest";
              }
              try {
                for (var O in S) {
                  R.setRequestHeader(O, S[O]);
                }
              } catch (Q) {}
              R.send((L.hasContent && L.data) || null);
              M = function (ac, W) {
                var X, V, U, aa, Z;
                try {
                  if (M && (W || R.readyState === 4)) {
                    M = undefined;
                    if (P) {
                      R.onreadystatechange = function () {};
                      n && delete k[P];
                    }
                    if (W) {
                      R.readyState !== 4 && R.abort();
                    } else {
                      X = R.status;
                      U = R.getAllResponseHeaders();
                      aa = {};
                      Z = R.responseXML;
                      Z && Z.documentElement && (aa.xml = Z);
                      try {
                        aa.text = R.responseText;
                      } catch (ab) {}
                      try {
                        V = R.statusText;
                      } catch (ab) {
                        V = "";
                      }
                      if (!X && L.isLocal && !L.crossDomain) {
                        X = aa.text ? 200 : 404;
                      } else {
                        if (X === 1223) {
                          X = 204;
                        }
                      }
                    }
                  }
                } catch (Y) {
                  !W && N(-1, Y);
                }
                aa && N(X, V, aa, U);
              };
              if (!L.async) {
                M();
              } else {
                if (R.readyState === 4) {
                  setTimeout(M, 0);
                } else {
                  P = ++C;
                  if (n) {
                    if (!k) {
                      k = {};
                      baidu.dom(window).on("unload", n);
                    }
                    k[P] = M;
                  }
                  R.onreadystatechange = M;
                }
              }
            },
            abort: function () {
              M && M(0, 1);
            },
          };
        }
      });
    }
  })();
  baidu.createChain(
    "fn",
    function (a) {
      return new baidu.fn.$Fn(~"function|string".indexOf(baidu.type(a)) ? a : function () {});
    },
    function (a) {
      this.fn = a;
    },
  );
  baidu.fn.blank = function () {};
  baidu.ajax.request = function (f, j) {
    var d = j || {},
      q = d.data || "",
      g = !(d.async === false),
      e = d.username || "",
      a = d.password || "",
      c = (d.method || "GET").toUpperCase(),
      b = d.headers || {},
      i = d.timeout || 0,
      k = {},
      n,
      r,
      h;
    function m() {
      if (h.readyState == 4) {
        try {
          var t = h.status;
        } catch (s) {
          p("failure");
          return;
        }
        p(t);
        if ((t >= 200 && t < 300) || t == 304 || t == 1223) {
          p("success");
        } else {
          p("failure");
        }
        window.setTimeout(function () {
          h.onreadystatechange = baidu.fn.blank;
          if (g) {
            h = null;
          }
        }, 0);
      }
    }
    function l() {
      if (window.ActiveXObject) {
        try {
          return new ActiveXObject("Msxml2.XMLHTTP");
        } catch (s) {
          try {
            return new ActiveXObject("Microsoft.XMLHTTP");
          } catch (s) {}
        }
      }
      if (window.XMLHttpRequest) {
        return new XMLHttpRequest();
      }
    }
    function p(u) {
      u = "on" + u;
      var t = k[u],
        v = baidu.ajax[u];
      if (t) {
        if (n) {
          clearTimeout(n);
        }
        if (u != "onsuccess") {
          t(h);
        } else {
          try {
            h.responseText;
          } catch (s) {
            return t(h);
          }
          t(h, h.responseText);
        }
      } else {
        if (v) {
          if (u == "onsuccess") {
            return;
          }
          v(h);
        }
      }
    }
    for (r in d) {
      k[r] = d[r];
    }
    b["X-Requested-With"] = "XMLHttpRequest";
    try {
      h = l();
      if (c == "GET") {
        if (q) {
          f += (f.indexOf("?") >= 0 ? "&" : "?") + q;
          q = null;
        }
        if (d["noCache"]) {
          f += (f.indexOf("?") >= 0 ? "&" : "?") + "b" + +new Date() + "=1";
        }
      }
      if (e) {
        h.open(c, f, g, e, a);
      } else {
        h.open(c, f, g);
      }
      if (g) {
        h.onreadystatechange = m;
      }
      if (c == "POST") {
        h.setRequestHeader("Content-Type", b["Content-Type"] || "application/x-www-form-urlencoded");
      }
      for (r in b) {
        if (b.hasOwnProperty(r)) {
          h.setRequestHeader(r, b[r]);
        }
      }
      p("beforerequest");
      if (i) {
        n = setTimeout(function () {
          h.onreadystatechange = baidu.fn.blank;
          h.abort();
          p("timeout");
        }, i);
      }
      h.send(q);
      if (!g) {
        m();
      }
    } catch (o) {
      p("failure");
    }
    return h;
  };
  baidu.url = baidu.url || {};
  baidu.url.escapeSymbol = function (a) {
    return String(a).replace(/[#%&+=\/\\\ \　\f\r\n\t]/g, function (b) {
      return "%" + (256 + b.charCodeAt()).toString(16).substring(1).toUpperCase();
    });
  };
  baidu.ajax.form = function (a, c) {
    c = c || {};
    var f = a.elements,
      n = f.length,
      b = a.getAttribute("method"),
      e = a.getAttribute("action"),
      t =
        c.replacer ||
        function (u, i) {
          return u;
        },
      q = {},
      s = [],
      l,
      p,
      r,
      m,
      d,
      g,
      h,
      k,
      j;
    function o(i, u) {
      s.push(i + "=" + u);
    }
    for (l in c) {
      if (c.hasOwnProperty(l)) {
        q[l] = c[l];
      }
    }
    for (l = 0; l < n; l++) {
      p = f[l];
      m = p.name;
      if (!p.disabled && m) {
        r = p.type;
        d = baidu.url.escapeSymbol(p.value);
        switch (r) {
          case "radio":
          case "checkbox":
            if (!p.checked) {
              break;
            }
          case "textarea":
          case "text":
          case "password":
          case "hidden":
          case "select-one":
            o(m, t(d, m));
            break;
          case "select-multiple":
            g = p.options;
            k = g.length;
            for (h = 0; h < k; h++) {
              j = g[h];
              if (j.selected) {
                o(m, t(j.value, m));
              }
            }
            break;
        }
      }
    }
    q.data = s.join("&");
    q.method = a.getAttribute("method") || "GET";
    return baidu.ajax.request(e, q);
  };
  baidu.ajax.get = function (b, a) {
    return baidu.ajax.request(b, { onsuccess: a });
  };
  baidu.ajax.post = function (b, c, a) {
    return baidu.ajax.request(b, { onsuccess: a, method: "POST", data: c });
  };
  baidu.array.extend({
    contains: function (a) {
      return this.indexOf(a) > -1;
    },
  });
  baidu.each = function (b, f, e) {
    var d, g, c, a;
    if (typeof f == "function" && b) {
      if (typeof b.length == "number") {
        for (d = 0, g = b.length; d < g; d++) {
          c = b[d] || (b.charAt && b.charAt(d));
          a = f.call(e || c, d, c, b);
          if (a === false || a == "break") {
            break;
          }
        }
      } else {
        if (typeof b == "number") {
          for (d = 0; d < b; d++) {
            a = f.call(e || d, d, d, d);
            if (a === false || a == "break") {
              break;
            }
          }
        } else {
          if (typeof b == "object") {
            for (d in b) {
              if (b.hasOwnProperty(d)) {
                a = f.call(e || b[d], d, b[d], b);
                if (a === false || a == "break") {
                  break;
                }
              }
            }
          }
        }
      }
    }
    return b;
  };
  baidu.array.extend({
    empty: function () {
      this.length = 0;
      return this;
    },
  });
  Array.prototype.every = function (c, b) {
    baidu.check("function(,.+)?", "baidu.array.every");
    var a, d;
    for (a = 0, d = this.length; a < d; a++) {
      if (!c.call(b || this, this[a], a, this)) {
        return false;
      }
    }
    return true;
  };
  baidu.array.every = function (c, b, a) {
    return baidu.isArray(c) ? c.every(b, a) : c;
  };
  Array.prototype.filter = function (func, thisArg) {
    try {
      if (!((typeof func === "Function" || typeof func === "function") && this)) {
        throw new TypeError();
      }
      var len = this.length >>> 0,
        res = new Array(len),
        t = this,
        c = 0,
        i = -1;
      if (thisArg === undefined) {
        while (++i !== len) {
          if (i in this) {
            if (func(t[i], i, t)) {
              res[c++] = t[i];
            }
          }
        }
      } else {
        while (++i !== len) {
          if (i in this) {
            if (func.call(thisArg, t[i], i, t)) {
              res[c++] = t[i];
            }
          }
        }
      }
      res.length = c;
      return res;
    } catch (e) {}
  };
  baidu.array.filter = function (c, b, a) {
    return baidu.isArray(c) ? c.filter(b, a) : [];
  };
  baidu.array.extend({
    find: function (b) {
      var a,
        c,
        d = this.length;
      if (baidu.type(b) == "function") {
        for (a = 0; a < d; a++) {
          c = this[a];
          if (b.call(this, c, a, this) === true) {
            return c;
          }
        }
      }
      return null;
    },
  });
  baidu.array.extend({
    hash: function (b) {
      var a = {},
        d = b && b.length,
        c,
        e;
      for (c = 0, e = this.length; c < e; c++) {
        a[this[c]] = d && d > c ? b[c] : true;
      }
      return a;
    },
  });
  baidu.array.extend({
    lastIndexOf: function (b, c) {
      baidu.check(".+(,number)?", "baidu.array.lastIndexOf");
      var a = this.length;
      (!(c = c | 0) || c >= a) && (c = a - 1);
      c < 0 && (c += a);
      for (; c >= 0; c--) {
        if (c in this && this[c] === b) {
          return c;
        }
      }
      return -1;
    },
  });
  Array.prototype.map = function (callback) {
    var T, A, k;
    if (this == null) {
      throw new TypeError("this is null or not defined");
    }
    var O = Object(this);
    var len = O.length >>> 0;
    if (typeof callback !== "function") {
      throw new TypeError(callback + " is not a function");
    }
    if (arguments.length > 1) {
      T = arguments[1];
    }
    A = new Array(len);
    k = 0;
    while (k < len) {
      var kValue, mappedValue;
      if (k in O) {
        kValue = O[k];
        mappedValue = callback.call(T, kValue, k, O);
        A[k] = mappedValue;
      }
      k++;
    }
    return A;
  };
  baidu.array.map = function (c, b, a) {
    return baidu.isArray(c) ? c.map(b, a) : c;
  };
  Array.prototype.reduce = function (b, c) {
    baidu.check("function(,.+)?", "baidu.array.reduce");
    var a = 0,
      e = this.length,
      d = false;
    if (typeof c == "undefined") {
      c = this[a++];
      if (typeof c == "undefined") {
        return;
      }
    }
    for (; a < e; a++) {
      c = b(c, this[a], a, this);
    }
    return c;
  };
  baidu.array.reduce = function (c, a, b) {
    return baidu.isArray(c) ? c.reduce(a, b) : c;
  };
  baidu.array.extend({
    remove: function (a) {
      var b = this.length;
      while (b--) {
        if (this[b] === a) {
          this.splice(b, 1);
        }
      }
      return this;
    },
  });
  baidu.array.extend({
    removeAt: function (a) {
      baidu.check("number", "baidu.array.removeAt");
      return this.splice(a, 1)[0];
    },
  });
  Array.prototype.some = function (c, b) {
    baidu.check("function(,.+)?", "baidu.array.some");
    var a, d;
    for (a = 0, d = this.length; a < d; a++) {
      if (c.call(b || this, this[a], a, this)) {
        return true;
      }
    }
    return false;
  };
  baidu.array.some = function (c, b, a) {
    return c.some(b, a);
  };
  baidu.createChain(
    "async",
    function (a) {
      return typeof a === "string" ? new baidu.async.$Async(a) : new baidu.async.$Async();
    },
    function (a) {
      this.url = a;
    },
  );
  baidu.object.extend = baidu.extend;
  baidu.lang.isFunction = baidu.isFunction;
  baidu.async._isDeferred = function (b) {
    var a = baidu.lang.isFunction;
    return b && a(b.success) && a(b.then) && a(b.fail) && a(b.cancel);
  };
  baidu.async.extend({
    Deferred: function () {
      baidu.async.Deferred.apply(this, arguments);
      return this;
    },
  });
  baidu.async.Deferred = function () {
    var b = this;
    baidu.extend(b, {
      _fired: 0,
      _firing: 0,
      _cancelled: 0,
      _resolveChain: [],
      _rejectChain: [],
      _result: [],
      _isError: 0,
    });
    function a() {
      if (b._cancelled || b._firing) {
        return;
      }
      if (b._nextDeferred) {
        b._nextDeferred.then(b._resolveChain[0], b._rejectChain[0]);
        return;
      }
      b._firing = 1;
      var f = b._isError ? b._rejectChain : b._resolveChain,
        c = b._result[b._isError ? 1 : 0];
      while (f[0] && !b._cancelled) {
        try {
          var d = f.shift().call(b, c);
          if (baidu.async._isDeferred(d)) {
            b._nextDeferred = d;
            [].push.apply(d._resolveChain, b._resolveChain);
            [].push.apply(d._rejectChain, b._rejectChain);
            f = b._resolveChain = [];
            b._rejectChain = [];
          }
        } catch (e) {
          throw e;
        } finally {
          b._fired = 1;
          b._firing = 0;
        }
      }
    }
    b.resolve = b.fireSuccess = function (c) {
      b._result[0] = c;
      a();
      return b;
    };
    b.reject = b.fireFail = function (c) {
      b._result[1] = c;
      b._isError = 1;
      a();
      return b;
    };
    b.then = function (c, d) {
      b._resolveChain.push(c);
      b._rejectChain.push(d);
      if (b._fired) {
        a();
      }
      return b;
    };
    b.success = function (c) {
      return b.then(c, baidu.fn.blank);
    };
    b.fail = function (c) {
      return b.then(baidu.fn.blank, c);
    };
    b.cancel = function () {
      b._cancelled = 1;
    };
  };
  baidu.async.extend({
    get: function () {
      var b = this.url;
      var a = new baidu.async.Deferred();
      baidu.ajax.request(b, {
        onsuccess: function (d, c) {
          a.resolve({ xhr: d, responseText: c });
        },
        onfailure: function (c) {
          a.reject({ xhr: c });
        },
      });
      return a;
    },
  });
  baidu.async.extend({
    post: function (c) {
      var b = this.url;
      var a = new baidu.async.Deferred();
      baidu.ajax.request(b, {
        method: "POST",
        data: c,
        onsuccess: function (e, d) {
          a.resolve({ xhr: e, responseText: d });
        },
        onfailure: function (d) {
          a.reject({ xhr: d });
        },
      });
      return a;
    },
  });
  baidu.async.when = function (c, b, d) {
    if (baidu.async._isDeferred(c)) {
      c.then(b, d);
      return c;
    }
    var a = new baidu.async.Deferred();
    a.then(b, d).resolve(c);
    return a;
  };
  baidu.base = baidu.base || {};
  baidu.cookie = baidu.cookie || {};
  baidu.cookie._isValidKey = function (a) {
    return new RegExp('^[^\\x00-\\x20\\x7f\\(\\)<>@,;:\\\\\\"\\[\\]\\?=\\{\\}\\/\\u0080-\\uffff]+\x24').test(a);
  };
  baidu.cookie.getRaw = function (b) {
    if (baidu.cookie._isValidKey(b)) {
      var c = new RegExp("(^| )" + b + "=([^;]*)(;|\x24)"),
        a = c.exec(document.cookie);
      if (a) {
        return a[2] || null;
      }
    }
    return null;
  };
  baidu.cookie.get = function (a) {
    var b = baidu.cookie.getRaw(a);
    if ("string" == typeof b) {
      b = decodeURIComponent(b);
      return b;
    }
    return null;
  };
  baidu.cookie.setRaw = function (c, d, b) {
    if (!baidu.cookie._isValidKey(c)) {
      return;
    }
    b = b || {};
    var a = b.expires;
    if ("number" == typeof b.expires) {
      a = new Date();
      a.setTime(a.getTime() + b.expires);
    }
    document.cookie =
      c +
      "=" +
      d +
      (b.path ? "; path=" + b.path : "") +
      (a ? "; expires=" + a.toGMTString() : "") +
      (b.domain ? "; domain=" + b.domain : "") +
      (b.secure ? "; secure" : "");
  };
  baidu.cookie.remove = function (b, a) {
    a = a || {};
    a.expires = new Date(0);
    baidu.cookie.setRaw(b, "", a);
  };
  baidu.cookie.set = function (b, c, a) {
    baidu.cookie.setRaw(b, encodeURIComponent(c), a);
  };
  baidu.object.isEmpty = function (c) {
    var a = true;
    if ("[object Array]" === Object.prototype.toString.call(c)) {
      a = !c.length;
    } else {
      c = new Object(c);
      for (var b in c) {
        return false;
      }
    }
    return a;
  };
  baidu._util_.cleanData = function (d) {
    var a;
    for (var b = 0, c; (c = d[b]); b++) {
      a = baidu.id(c, "get");
      if (!a) {
        continue;
      }
      baidu._util_.eventBase.removeAll(c);
      baidu.id(c, "remove");
    }
  };
  baidu.json = baidu.json || {};
  baidu.json.parse = function (a) {
    return new Function("return (" + a + ")")();
  };
  (function () {
    var f = /^(?:\{.*\}|\[.*\])$/,
      a = /([A-Z])/g,
      c = /^-ms-/,
      e = /-([\da-z])/gi,
      b = function (h, i) {
        return (i + "").toUpperCase();
      };
    baidu.extend(baidu._util_, {
      camelCase: function (h) {
        return h.replace(c, "ms-").replace(e, b);
      },
    });
    baidu.extend(baidu._util_, {
      cache: {},
      deletedIds: [],
      uuid: 0,
      expando: "baidu" + ("2.0.0" + Math.random()).replace(/\D/g, ""),
      noData: { embed: true, object: "clsid:D27CDB6E-AE6D-11cf-96B8-444553540000", applet: true },
      hasData: function (h) {
        h = h.nodeType ? baidu._util_.cache[h[baidu._util_.expando]] : h[baidu._util_.expando];
        return !!h && !d(h);
      },
      data: function (k, i, m, l) {
        if (!baidu._util_.acceptData(k)) {
          return;
        }
        var n,
          p,
          q = baidu._util_.expando,
          o = typeof i === "string",
          r = k.nodeType,
          h = r ? baidu._util_.cache : k,
          j = r ? k[q] : k[q] && q;
        if ((!j || !h[j] || (!l && !h[j].data)) && o && m === undefined) {
          return;
        }
        if (!j) {
          if (r) {
            k[q] = j = baidu._util_.deletedIds.pop() || ++baidu._util_.uuid;
          } else {
            j = q;
          }
        }
        if (!h[j]) {
          h[j] = {};
          if (!r) {
            h[j].toJSON = function () {};
          }
        }
        if (typeof i === "object" || typeof i === "function") {
          if (l) {
            h[j] = baidu.extend(h[j], i);
          } else {
            h[j].data = baidu.extend(h[j].data, i);
          }
        }
        n = h[j];
        if (!l) {
          if (!n.data) {
            n.data = {};
          }
          n = n.data;
        }
        if (m !== undefined) {
          n[baidu._util_.camelCase(i)] = m;
        }
        if (o) {
          p = n[i];
          if (p == null) {
            p = n[baidu._util_.camelCase(i)];
          }
        } else {
          p = n;
        }
        return p;
      },
      removeData: function (m, j, n) {
        if (!baidu._util_.acceptData(m)) {
          return;
        }
        var q,
          p,
          o,
          r = m.nodeType,
          h = r ? baidu._util_.cache : m,
          k = r ? m[baidu._util_.expando] : baidu._util_.expando;
        if (!h[k]) {
          return;
        }
        if (j) {
          q = n ? h[k] : h[k].data;
          if (q) {
            if (!baidu.isArray(j)) {
              if (j in q) {
                j = [j];
              } else {
                j = baidu._util_.camelCase(j);
                if (j in q) {
                  j = [j];
                } else {
                  j = j.split(" ");
                }
              }
            }
            for (p = 0, o = j.length; p < o; p++) {
              delete q[j[p]];
            }
            if (!(n ? d : baidu.object.isEmpty)(q)) {
              return;
            }
          }
        }
        if (!n) {
          delete h[k].data;
          if (!d(h[k])) {
            return;
          }
        }
        if (r) {
          baidu._util_.cleanData([m], true);
        } else {
          if (baidu.support.deleteExpando || h != h.window) {
            delete h[k];
          } else {
            h[k] = null;
          }
        }
      },
      _data: function (i, h, j) {
        return baidu._util_.data(i, h, j, true);
      },
      acceptData: function (i) {
        var h = i.nodeName && baidu._util_.noData[i.nodeName.toLowerCase()];
        return !h || (h !== true && i.getAttribute("classid") === h);
      },
    });
    function g(j, i, k) {
      if (k === undefined && j.nodeType === 1) {
        var h = "data-" + i.replace(a, "-$1").toLowerCase();
        k = j.getAttribute(h);
        if (typeof k === "string") {
          try {
            k =
              k === "true"
                ? true
                : k === "false"
                  ? false
                  : k === "null"
                    ? null
                    : baidu.isNumber(k)
                      ? +k
                      : f.test(k)
                        ? baidu.json.parse(k)
                        : k;
          } catch (l) {}
          baidu._util_.data(j, i, k);
        } else {
          k = undefined;
        }
      }
      return k;
    }
    function d(i) {
      var h;
      for (h in i) {
        if (h === "data" && baidu.object.isEmpty(i[h])) {
          continue;
        }
        if (h !== "toJSON") {
          return false;
        }
      }
      return true;
    }
  })();
  baidu.date = baidu.date || {};
  baidu.createChain("number", function (b) {
    var a = parseFloat(b),
      c = isNaN(a) ? a : b;
    ((clazz = typeof c === "number" ? Number : String), (pro = clazz.prototype));
    c = new clazz(c);
    baidu.forEach(baidu.number.$Number.prototype, function (e, d) {
      pro[d] || (c[d] = e);
    });
    return c;
  });
  baidu.number.extend({
    pad: function (c) {
      var d = this;
      var e = "",
        b = d < 0,
        a = String(Math.abs(d));
      if (a.length < c) {
        e = new Array(c - a.length + 1).join("0");
      }
      return (b ? "-" : "") + e + a;
    },
  });
  baidu.date.format = function (a, f) {
    if ("string" != typeof f) {
      return a.toString();
    }
    function d(l, k) {
      f = f.replace(l, k);
    }
    var b = baidu.number.pad,
      g = a.getFullYear(),
      e = a.getMonth() + 1,
      j = a.getDate(),
      h = a.getHours(),
      c = a.getMinutes(),
      i = a.getSeconds();
    d(/yyyy/g, b(g, 4));
    d(/yy/g, b(parseInt(g.toString().slice(2), 10), 2));
    d(/MM/g, b(e, 2));
    d(/M/g, e);
    d(/dd/g, b(j, 2));
    d(/d/g, j);
    d(/HH/g, b(h, 2));
    d(/H/g, h);
    d(/hh/g, b(h % 12, 2));
    d(/h/g, h % 12);
    d(/mm/g, b(c, 2));
    d(/m/g, c);
    d(/ss/g, b(i, 2));
    d(/s/g, i);
    return f;
  };
  baidu.date.parse = function (c) {
    var a = new RegExp("^\\d+(\\-|\\/)\\d+(\\-|\\/)\\d+\x24");
    if ("string" == typeof c) {
      if (a.test(c) || isNaN(Date.parse(c))) {
        var f = c.split(/ |T/),
          b = f.length > 1 ? f[1].split(/[^\d]/) : [0, 0, 0],
          e = f[0].split(/[^\d]/);
        return new Date(e[0] - 0, e[1] - 1, e[2] - 0, b[0] - 0, b[1] - 0, b[2] - 0);
      } else {
        return new Date(c);
      }
    }
    return new Date();
  };
  baidu.dom.createElements = (function () {
    var c = /<(\w+)/i,
      b = /<|&#?\w+;/,
      d = {
        area: [1, "<map>", "</map>"],
        col: [2, "<table><tbody></tbody><colgroup>", "</colgroup></table>"],
        legend: [1, "<fieldset>", "</fieldset>"],
        option: [1, "<select multiple='multiple'>", "</select>"],
        td: [3, "<table><tbody><tr>", "</tr></tbody></table>"],
        thead: [1, "<table>", "</table>"],
        tr: [2, "<table><tbody>", "</tbody></table>"],
        _default: [0, "", ""],
      };
    d.optgroup = d.option;
    d.tbody = d.tfoot = d.colgroup = d.caption = d.thead;
    d.th = d.td;
    function a(h, k) {
      var j = h.getElementsByTagName("SCRIPT"),
        f,
        e,
        g;
      for (f = j.length - 1; f >= 0; f--) {
        g = j[f];
        e = k.createElement("SCRIPT");
        g.id && (e.id = g.id);
        g.src && (e.src = g.src);
        g.type && (e.type = g.type);
        e[g.text ? "text" : "textContent"] = g.text || g.textContent;
        g.parentNode.replaceChild(e, g);
      }
    }
    return function (i, l) {
      baidu.isNumber(i) && (i = i.toString());
      l = l || document;
      var f,
        h,
        j,
        m = i,
        g = m.length,
        e = l.createElement("div"),
        k = l.createDocumentFragment(),
        o = [];
      if (baidu.isString(m)) {
        if (!b.test(m)) {
          o.push(l.createTextNode(m));
        } else {
          f = d[m.match(c)[1].toLowerCase()] || d._default;
          e.innerHTML = "<i>mz</i>" + f[1] + m + f[2];
          e.removeChild(e.firstChild);
          a(e, l);
          h = f[0];
          j = e;
          while (h--) {
            j = j.firstChild;
          }
          baidu.merge(o, j.childNodes);
          baidu.forEach(o, function (n) {
            k.appendChild(n);
          });
          e = j = null;
        }
      }
      e = null;
      return o;
    };
  })();
  baidu.dom.extend({
    add: function (c, d) {
      var b = baidu.array(this.get());
      switch (baidu.type(c)) {
        case "HTMLElement":
          b.push(c);
          break;
        case "$DOM":
        case "array":
          baidu.merge(b, c);
          break;
        case "string":
          baidu.merge(b, baidu.dom(c, d));
          break;
        default:
          if (typeof c == "object" && c.length) {
            baidu.merge(b, c);
          }
      }
      return baidu.dom(b.unique());
    },
  });
  baidu.dom.extend({
    addClass: function (b) {
      if (arguments.length <= 0) {
        return this;
      }
      switch (typeof b) {
        case "string":
          b = b.replace(/^\s+/g, "").replace(/\s+$/g, "").replace(/\s+/g, " ");
          var a = b.split(" ");
          baidu.forEach(this, function (e, c) {
            var f = "";
            if (e.className) {
              f = e.className;
            }
            for (var d = 0; d < a.length; d++) {
              if ((" " + f + " ").indexOf(" " + a[d] + " ") == -1) {
                f += " " + a[d];
              }
            }
            e.className = f.replace(/^\s+/g, "");
          });
          break;
        case "function":
          baidu.forEach(this, function (d, c) {
            baidu.dom(d).addClass(b.call(d, c, d.className));
          });
          break;
      }
      return this;
    },
  });
  baidu.dom.extend({
    getDocument: function () {
      if (this.size() <= 0) {
        return undefined;
      }
      var a = this[0];
      return a.nodeType == 9 ? a : a.ownerDocument || a.document;
    },
  });
  baidu.dom.extend({
    empty: function () {
      for (var a = 0, b; (b = this[a]); a++) {
        b.nodeType === 1 && baidu._util_.cleanData(b.getElementsByTagName("*"));
        while (b.firstChild) {
          b.removeChild(b.firstChild);
        }
      }
      return this;
    },
  });
  baidu.dom.extend({
    append: function () {
      baidu.check(
        "^(?:string|function|HTMLElement|\\$DOM)(?:,(?:string|array|HTMLElement|\\$DOM))*$",
        "baidu.dom.append",
      );
      baidu._util_.smartInsert(this, arguments, function (a) {
        this.nodeType === 1 && this.appendChild(a);
      });
      return this;
    },
  });
  baidu.dom.extend({
    html: function (j) {
      var d = baidu.dom,
        k = baidu._util_,
        g = this,
        h = false,
        m;
      if (this.size() <= 0) {
        switch (typeof j) {
          case "undefined":
            return undefined;
            break;
          default:
            return g;
            break;
        }
      }
      var c =
          "abbr|article|aside|audio|bdi|canvas|data|datalist|details|figcaption|figure|footer|" +
          "header|hgroup|mark|meter|nav|output|progress|section|summary|time|video",
        i = /<(?:script|style|link)/i,
        b = new RegExp("<(?:" + c + ")[\\s/>]", "i"),
        f = /^\s+/,
        a = /<(?!area|br|col|embed|hr|img|input|link|meta|param)(([\w:]+)[^>]*)\/>/gi,
        e = /<([\w:]+)/,
        l = {
          option: [1, "<select multiple='multiple'>", "</select>"],
          legend: [1, "<fieldset>", "</fieldset>"],
          thead: [1, "<table>", "</table>"],
          tr: [2, "<table><tbody>", "</tbody></table>"],
          td: [3, "<table><tbody><tr>", "</tr></tbody></table>"],
          col: [2, "<table><tbody></tbody><colgroup>", "</colgroup></table>"],
          area: [1, "<map>", "</map>"],
          _default: [0, "", ""],
        };
      l.optgroup = l.option;
      l.tbody = l.tfoot = l.colgroup = l.caption = l.thead;
      l.th = l.td;
      if (!baidu.support.htmlSerialize) {
        l._default = [1, "X<div>", "</div>"];
      }
      baidu.forEach(g, function (o, n) {
        if (m) {
          return;
        }
        var q = d(o);
        switch (typeof j) {
          case "undefined":
            m = o.nodeType === 1 ? o.innerHTML : undefined;
            return m;
            break;
          case "number":
            j = String(j);
          case "string":
            h = true;
            if (
              !i.test(j) &&
              (baidu.support.htmlSerialize || !b.test(j)) &&
              (baidu.support.leadingWhitespace || !f.test(j)) &&
              !l[(e.exec(j) || ["", ""])[1].toLowerCase()]
            ) {
              j = j.replace(a, "<$1></$2>");
              try {
                if (o.nodeType === 1) {
                  q.empty();
                  o.innerHTML = j;
                }
                o = 0;
              } catch (p) {}
            }
            if (o) {
              g.empty().append(j);
            }
            break;
          case "function":
            h = true;
            q.html(j.call(o, n, q.html()));
            break;
        }
      });
      return h ? g : m;
    },
  });
  baidu._util_.smartInsert = function (a, e, k) {
    if (e.length <= 0 || a.size() <= 0) {
      return;
    }
    if (baidu.type(e[0]) === "function") {
      var g = e[0],
        j;
      return baidu.forEach(a, function (m, i) {
        j = baidu.dom(m);
        e[0] = g.call(m, i, j.html());
        baidu._util_.smartInsert(j, e, k);
      });
    }
    var h = a.getDocument() || document,
      d = h.createDocumentFragment(),
      c = a.length - 1,
      f;
    for (var b = 0, l; (l = e[b]); b++) {
      if (l.nodeType) {
        d.appendChild(l);
      } else {
        baidu.forEach(~"string|number".indexOf(baidu.type(l)) ? baidu.dom.createElements(l, h) : l, function (i) {
          d.appendChild(i);
        });
      }
    }
    if (!(f = d.firstChild)) {
      return;
    }
    baidu.forEach(a, function (m, i) {
      k.call(
        m.nodeName.toLowerCase() === "table" && f.nodeName.toLowerCase() === "tr"
          ? m.tBodies[0] || m.appendChild(m.ownerDocument.createElement("tbody"))
          : m,
        i < c ? d.cloneNode(true) : d,
      );
    });
  };
  baidu.dom.extend({
    after: function () {
      baidu.check(
        "^(?:string|function|HTMLElement|\\$DOM)(?:,(?:string|array|HTMLElement|\\$DOM))*$",
        "baidu.dom.after",
      );
      var a = this[0] && this[0].parentNode,
        b = !a && [];
      baidu._util_.smartInsert(this, arguments, function (c) {
        a ? a.insertBefore(c, this.nextSibling) : baidu.merge(b, c.childNodes);
      });
      b && baidu.merge(this, b);
      return this;
    },
  });
  baidu.dom.extend({
    map: function (a) {
      baidu.check("function", "baidu.dom.map");
      var b = this,
        c = baidu.dom();
      baidu.forEach(this, function (e, d) {
        c[c.length++] = a.call(e, d, e, e);
      });
      return c;
    },
  });
  baidu._util_.isXML = function (b) {
    var a = (b ? b.ownerDocument || b : 0).documentElement;
    return a ? a.nodeName !== "HTML" : false;
  };
  baidu.dom.extend({
    clone: (function () {
      var d = baidu._util_.eventBase;
      function b(f) {
        return f.getElementsByTagName ? f.getElementsByTagName("*") : f.querySelectorAll ? f.querySelectorAll("*") : [];
      }
      function c(g, f) {
        f.clearAttributes && f.clearAttributes();
        f.mergeAttributes && f.mergeAttributes(g);
        switch (f.nodeName.toLowerCase()) {
          case "object":
            f.outerHTML = g.outerHTML;
            break;
          case "textarea":
          case "input":
            if (~"checked|radio".indexOf(g.type)) {
              g.checked && (f.defaultChecked = f.checked = g.checked);
              f.value !== g.value && (f.value = g.value);
            }
            f.defaultValue = g.defaultValue;
            break;
          case "options":
            f.selected = g.defaultSelected;
            break;
          case "script":
            f.text !== g.text && (f.text = g.text);
            break;
        }
        f[baidu.key] && f.removeAttribute(baidu.key);
      }
      function a(m, h) {
        if (h.nodeType !== 1 || !baidu.id(m, "get")) {
          return;
        }
        var f = d.get(m);
        for (var k in f) {
          for (var g = 0, l; (l = f[k][g]); g++) {
            d.add(h, k, l);
          }
        }
      }
      function e(m, n, k) {
        var j = m.cloneNode(true),
          g,
          h,
          f;
        if (
          (!baidu.support.noCloneEvent || !baidu.support.noCloneChecked) &&
          (m.nodeType === 1 || m.nodeType === 11) &&
          !baidu._util_.isXML(m)
        ) {
          c(m, j);
          g = b(m);
          h = b(j);
          f = g.length;
          for (var l = 0; l < f; l++) {
            h[l] && c(g[l], h[l]);
          }
        }
        if (n) {
          a(m, j);
          if (k) {
            g = b(m);
            h = b(j);
            f = g.length;
            for (var l = 0; l < f; l++) {
              a(g[l], h[l]);
            }
          }
        }
        return j;
      }
      return function (g, f) {
        g = !!g;
        f = !!f;
        return this.map(function () {
          return e(this, g, f);
        });
      };
    })(),
  });
  baidu._util_.smartInsertTo = function (a, d, f, h) {
    var g = baidu.dom(d),
      c = g[0],
      e;
    if (h && c && (!c.parentNode || c.parentNode.nodeType === 11)) {
      h = h === "before";
      e = baidu.merge(h ? a : g, !h ? a : g);
      if (a !== e) {
        a.length = 0;
        baidu.merge(a, e);
      }
    } else {
      for (var b = 0, j; (j = g[b]); b++) {
        baidu._util_.smartInsert(baidu.dom(j), b > 0 ? a.clone(true) : a, f);
      }
    }
  };
  baidu.dom.extend({
    appendTo: function (a) {
      baidu.check("^(?:string|HTMLElement|\\$DOM)$", "baidu.dom.appendTo");
      baidu._util_.smartInsertTo(this, a, function (b) {
        this.appendChild(b);
      });
      return this;
    },
  });
  baidu.extend(baidu._util_, {
    rfocusable: /^(?:button|input|object|select|textarea)$/i,
    rclickable: /^a(?:rea)?$/i,
    rboolean:
      /^(?:autofocus|autoplay|async|checked|controls|defer|disabled|hidden|loop|multiple|open|readonly|required|scoped|selected)$/i,
    propFix: {
      tabindex: "tabIndex",
      readonly: "readOnly",
      for: "htmlFor",
      class: "className",
      maxlength: "maxLength",
      cellspacing: "cellSpacing",
      cellpadding: "cellPadding",
      rowspan: "rowSpan",
      colspan: "colSpan",
      usemap: "useMap",
      frameborder: "frameBorder",
      contenteditable: "contentEditable",
    },
    propHooks: {
      tabIndex: {
        get: function (c) {
          var a = baidu._util_;
          var b = c.getAttributeNode("tabindex");
          return b && b.specified
            ? parseInt(b.value, 10)
            : a.rfocusable.test(c.nodeName) || (a.rclickable.test(c.nodeName) && c.href)
              ? 0
              : undefined;
        },
      },
    },
  });
  if (!baidu.support.enctype) {
    var bu = baidu._util_;
    bu.propFix.enctype = "encoding";
  }
  if (!baidu.support.optSelected) {
    var bu = baidu._util_;
    bu.propHooks.selected = baidu.extend(bu.propHooks.selected, {
      get: function (b) {
        var a = b.parentNode;
        if (a) {
          a.selectedIndex;
          if (a.parentNode) {
            a.parentNode.selectedIndex;
          }
        }
        return null;
      },
    });
  }
  baidu.extend(baidu, {
    _error: function (a) {
      throw new Error(a);
    },
    _nodeName: function (b, a) {
      return b.nodeName && b.nodeName.toUpperCase() === a.toUpperCase();
    },
  });
  baidu.extend(baidu._util_, {
    rfocusable: /^(?:button|input|object|select|textarea)$/i,
    rtype: /^(?:button|input)$/i,
    rclickable: /^a(?:rea)?$/i,
    nodeHook: {},
    attrHooks: {
      type: {
        set: function (b, c) {
          var a = baidu._util_;
          if (a.rtype.test(b.nodeName) && b.parentNode) {
            baidu._error("type property can't be changed");
          } else {
            if (!baidu.support.radioValue && c === "radio" && baidu._nodeName(b, "input")) {
              var d = b.value;
              b.setAttribute("type", c);
              if (d) {
                b.value = d;
              }
              return c;
            }
          }
        },
      },
      value: {
        get: function (c, b) {
          var a = baidu._util_;
          if (a.nodeHook && baidu._nodeName(c, "button")) {
            return a.nodeHook.get(c, b);
          }
          return b in c ? c.value : null;
        },
        set: function (b, c, a) {
          if (bu.nodeHook && baidu._nodeName(b, "button")) {
            return bu.nodeHook.set(b, c, a);
          }
          b.value = c;
        },
      },
    },
    boolHook: {
      get: function (b, a) {
        var d,
          c = baidu(b).prop(a);
        return c === true || (typeof c !== "boolean" && (d = b.getAttributeNode(a)) && d.nodeValue !== false)
          ? a.toLowerCase()
          : undefined;
      },
      set: function (b, d, a) {
        var c;
        if (d === false) {
          baidu(b).removeAttr(a);
        } else {
          c = baidu._util_.propFix[a] || a;
          if (c in b) {
            b[c] = true;
          }
          b.setAttribute(a, a.toLowerCase());
        }
        return a;
      },
    },
  });
  baidu._util_.attrHooks.tabindex = baidu._util_.propHooks.tabIndex;
  if (!baidu.support.getSetAttribute) {
    var bu = baidu._util_,
      fixSpecified = { name: true, id: true, coords: true };
    bu.nodeHook = {
      get: function (c, b) {
        var a;
        a = c.getAttributeNode(b);
        return a && (fixSpecified[b] ? a.nodeValue !== "" : a.specified) ? a.nodeValue : undefined;
      },
      set: function (c, d, b) {
        var a = c.getAttributeNode(b);
        if (!a) {
          a = document.createAttribute(b);
          c.setAttributeNode(a);
        }
        return (a.nodeValue = d + "");
      },
    };
    bu.attrHooks.tabindex.set = bu.nodeHook.set;
    baidu.forEach(["width", "height"], function (a) {
      bu.attrHooks[a] = baidu.extend(bu.attrHooks[a], {
        set: function (b, c) {
          if (c === "") {
            b.setAttribute(a, "auto");
            return c;
          }
        },
      });
    });
    bu.attrHooks.contenteditable = {
      get: bu.nodeHook.get,
      set: function (b, c, a) {
        if (c === "") {
          c = "false";
        }
        bu.nodeHook.set(b, c, a);
      },
    };
  }
  if (!baidu.support.hrefNormalized) {
    var bu = baidu._util_;
    baidu.forEach(["href", "src", "width", "height"], function (a) {
      bu.attrHooks[a] = baidu.extend(bu.attrHooks[a], {
        get: function (c) {
          var b = c.getAttribute(a, 2);
          return b === null ? undefined : b;
        },
      });
    });
  }
  if (!baidu.support.style) {
    var bu = baidu._util_;
    bu.attrHooks.style = {
      get: function (a) {
        return a.style.cssText.toLowerCase() || undefined;
      },
      set: function (a, b) {
        return (a.style.cssText = "" + b);
      },
    };
  }
  baidu.dom.extend({
    prop: function (b, e) {
      if (arguments.length <= 0 || typeof b === "function") {
        return this;
      }
      var a,
        d = this,
        c = false;
      if (this.size() <= 0) {
        if (b && e) {
          return d;
        } else {
          if (b && !e) {
            return undefined;
          } else {
            return d;
          }
        }
      }
      baidu.forEach(this, function (m, i) {
        if (a) {
          return;
        }
        var j,
          l,
          f,
          h = baidu.dom,
          k = baidu._util_,
          g = m.nodeType;
        if (!m || g === 3 || g === 8 || g === 2) {
          return;
        }
        f = g !== 1 || !baidu._util_.isXML(m);
        if (f) {
          b = k.propFix[b] || b;
          l = k.propHooks[b];
        }
        switch (typeof b) {
          case "string":
            if (typeof e === "undefined") {
              if (l && "get" in l && (j = l.get(m, b)) !== null) {
                a = j;
              } else {
                a = m[b];
              }
            } else {
              if (typeof e === "function") {
                c = true;
                var n = h(m);
                n.prop(b, e.call(m, i, n.prop(b)));
              } else {
                c = true;
                if (l && "set" in l && (j = l.set(m, e, b)) !== undefined) {
                  return j;
                } else {
                  m[b] = e;
                }
              }
            }
            break;
          case "object":
            c = true;
            var n = h(m);
            for (key in b) {
              n.prop(key, b[key]);
            }
            break;
          default:
            a = d;
            break;
        }
      });
      return c ? this : a;
    },
  });
  baidu.makeArray = function (c, b) {
    var a = b || [];
    if (!c) {
      return a;
    }
    c.length == null || ~"string|function|regexp".indexOf(baidu.type(c))
      ? Array.prototype.push.call(a, c)
      : baidu.merge(a, c);
    return a;
  };
  baidu.extend(baidu._util_, {
    nodeName: function (b, a) {
      return b.nodeName && b.nodeName.toUpperCase() === a.toUpperCase();
    },
    valHooks: {
      option: {
        get: function (a) {
          var b = a.attributes.value;
          return !b || b.specified ? a.value : a.text;
        },
      },
      select: {
        get: function (a) {
          var g,
            b,
            f,
            d,
            e = a.selectedIndex,
            h = [],
            j = a.options,
            c = a.type === "select-one";
          if (e < 0) {
            return null;
          }
          b = c ? e : 0;
          f = c ? e + 1 : j.length;
          for (; b < f; b++) {
            d = j[b];
            if (
              d.selected &&
              (baidu.support.optDisabled ? !d.disabled : d.getAttribute("disabled") === null) &&
              (!d.parentNode.disabled || !baidu._util_.nodeName(d.parentNode, "optgroup"))
            ) {
              g = baidu.dom(d).val();
              if (c) {
                return g;
              }
              h.push(g);
            }
          }
          if (c && !h.length && j.length) {
            return baidu(j[e]).val();
          }
          return h;
        },
        set: function (b, c) {
          var a = baidu.makeArray(c);
          baidu(b)
            .find("option")
            .each(function () {
              this.selected = baidu.array(a).indexOf(baidu(this).val()) >= 0;
            });
          if (!a.length) {
            b.selectedIndex = -1;
          }
          return a;
        },
      },
    },
  });
  if (!baidu.support.getSetAttribute) {
    var fixSpecified = { name: true, id: true, coords: true };
    baidu._util_.valHooks.button = {
      get: function (c, b) {
        var a;
        a = c.getAttributeNode(b);
        return a && (fixSpecified[b] ? a.value !== "" : a.specified) ? a.value : undefined;
      },
      set: function (c, d, b) {
        var a = c.getAttributeNode(b);
        if (!a) {
          a = document.createAttribute(b);
          c.setAttributeNode(a);
        }
        return (a.value = d + "");
      },
    };
  }
  if (!baidu.support.checkOn) {
    baidu.forEach(["radio", "checkbox"], function () {
      baidu._util_.valHooks[this] = {
        get: function (a) {
          return a.getAttribute("value") === null ? "on" : a.value;
        },
      };
    });
  }
  baidu.forEach(["radio", "checkbox"], function (a) {
    baidu._util_.valHooks[a] = baidu.extend(baidu._util_.valHooks[a], {
      set: function (b, c) {
        if (baidu.isArray(c)) {
          return (b.checked = baidu.array(c).indexOf(baidu(b).val()) >= 0);
        }
      },
    });
  });
  baidu.dom.extend({
    val: function (f) {
      var e = baidu.dom,
        b = baidu._util_,
        d = this,
        c = false,
        a;
      if (this.size() <= 0) {
        switch (typeof f) {
          case "undefined":
            return undefined;
            break;
          default:
            return d;
            break;
        }
      }
      baidu.forEach(d, function (i, h) {
        var g,
          j = /\r/g;
        if (a) {
          return;
        }
        switch (typeof f) {
          case "undefined":
            if (i) {
              g = b.valHooks[i.type] || b.valHooks[i.nodeName.toLowerCase()];
              if (g && "get" in g && (a = g.get(i, "value")) !== undefined) {
                return a;
              }
              a = i.value;
              return typeof a === "string" ? a.replace(j, "") : a == null ? "" : a;
            }
            return a;
            break;
          case "function":
            c = true;
            var k = e(i);
            k.val(f.call(i, h, k.val()));
            break;
          default:
            c = true;
            if (i.nodeType !== 1) {
              return;
            }
            if (f == null) {
              f = "";
            } else {
              if (typeof f === "number") {
                f += "";
              } else {
                if (baidu.isArray(f)) {
                  f = baidu.forEach(f, function (m, l) {
                    return m == null ? "" : m + "";
                  });
                }
              }
            }
            g = b.valHooks[i.type] || b.valHooks[i.nodeName.toLowerCase()];
            if (!g || !("set" in g) || g.set(i, f, "value") === undefined) {
              i.value = f;
            }
            break;
        }
      });
      return c ? d : a;
    },
  });
  baidu._util_.access = function (b, d, e) {
    if (this.size() <= 0) {
      return this;
    }
    switch (baidu.type(b)) {
      case "string":
        if (d === undefined) {
          return e.call(this, this[0], b);
        } else {
          for (var a = 0, c; (c = this[a]); a++) {
            e.call(this, c, b, baidu.type(d) === "function" ? d.call(c, a, e.call(this, c, b)) : d);
          }
        }
        break;
      case "object":
        for (var a in b) {
          baidu._util_.access.call(this, a, b[a], e);
        }
        break;
    }
    return this;
  };
  baidu.dom.extend({
    getComputedStyle: function (b) {
      var c = this[0].ownerDocument.defaultView,
        a = c && c.getComputedStyle && c.getComputedStyle(this[0], null),
        d = a ? a.getPropertyValue(b) || a[b] : "";
      return d || this[0].style[b];
    },
  });
  baidu.dom.extend({
    getCurrentStyle: (function () {
      var a = document.documentElement.currentStyle
        ? function (b) {
            return this[0].currentStyle ? this[0].currentStyle[b] : this[0].style[b];
          }
        : function (b) {
            return this.getComputedStyle(b);
          };
      return function (b) {
        return a.call(this, b);
      };
    })(),
  });
  baidu._util_.getWidthOrHeight = (function () {
    var a = {},
      c = { position: "absolute", visibility: "hidden", display: "block" };
    function b(f, d) {
      var g = {};
      for (var e in d) {
        g[e] = f.style[e];
        f.style[e] = d[e];
      }
      return g;
    }
    baidu.forEach(["Width", "Height"], function (d) {
      var e = { Width: ["Right", "Left"], Height: ["Top", "Bottom"] }[d];
      a["get" + d] = function (k, f) {
        var j = baidu.dom(k),
          i = k["offset" + d],
          g = i === 0 && b(k, c),
          h = "padding|border";
        g && (i = k["offset" + d]);
        f &&
          baidu.forEach(f.split("|"), function (l) {
            if (!~h.indexOf(l)) {
              i += parseFloat(j.getCurrentStyle(l + e[0])) || 0;
              i += parseFloat(j.getCurrentStyle(l + e[1])) || 0;
            } else {
              h = h.replace(new RegExp("\\|?" + l + "\\|?"), "");
            }
          });
        h &&
          baidu.forEach(h.split("|"), function (l) {
            i -= parseFloat(j.getCurrentStyle(l + e[0] + (l === "border" ? "Width" : ""))) || 0;
            i -= parseFloat(j.getCurrentStyle(l + e[1] + (l === "border" ? "Width" : ""))) || 0;
          });
        g && b(k, g);
        return i;
      };
    });
    return function (f, e, d) {
      return a[e === "width" ? "getWidth" : "getHeight"](f, d);
    };
  })();
  baidu.string.extend({
    toCamelCase: function () {
      var a = this.valueOf();
      if (a.indexOf("-") < 0 && a.indexOf("_") < 0) {
        return a;
      }
      return a.replace(/[-_][^-_]/g, function (b) {
        return b.charAt(1).toUpperCase();
      });
    },
  });
  baidu.dom.styleFixer = (function () {
    var e = /alpha\s*\(\s*opacity\s*=\s*(\d{1,3})/i,
      a = /^-?(?:\d*\.)?\d+(?!px)[^\d\s]+$/i,
      f = "fillOpacity,fontWeight,opacity,orphans,widows,zIndex,zoom",
      g = { float: baidu.support.cssFloat ? "cssFloat" : "styleFloat" },
      b = { fontWeight: { normal: 400, bold: 700, bolder: 700, lighter: 100 } },
      d = {
        opacity: {},
        width: {},
        height: {},
        fontWeight: {
          get: function (j, i) {
            var h = c.get(j, i);
            return b.fontWeight[h] || h;
          },
        },
      },
      c = {
        set: function (i, h, j) {
          i.style[h] = j;
        },
      };
    baidu.extend(
      d.opacity,
      baidu.support.opacity
        ? {
            get: function (j, i) {
              var h = baidu.dom(j).getCurrentStyle(i);
              return h === "" ? "1" : h;
            },
          }
        : {
            get: function (h) {
              return e.test((h.currentStyle || h.style).filter || "") ? parseFloat(RegExp.$1) / 100 : "1";
            },
            set: function (l, i, k) {
              var j = (l.currentStyle || l.style).filter || "",
                h = k * 100;
              l.style.zoom = 1;
              l.style.filter = e.test(j)
                ? j.replace(e, "Alpha(opacity=" + h)
                : j + " progid:dximagetransform.microsoft.Alpha(opacity=" + h + ")";
            },
          },
    );
    baidu.forEach(["width", "height"], function (h) {
      d[h] = {
        get: function (i) {
          return baidu._util_.getWidthOrHeight(i, h) + "px";
        },
        set: function (j, i, k) {
          baidu.type(k) === "number" && k < 0 && (k = 0);
          c.set(j, i, k);
        },
      };
    });
    baidu.extend(
      c,
      document.documentElement.currentStyle
        ? {
            get: function (j, i) {
              var k = baidu.dom(j).getCurrentStyle(i),
                h;
              if (a.test(k)) {
                h = j.style.left;
                j.style.left = i === "fontSize" ? "1em" : k;
                k = j.style.pixelLeft + "px";
                j.style.left = h;
              }
              return k;
            },
          }
        : {
            get: function (i, h) {
              return baidu.dom(i).getCurrentStyle(h);
            },
          },
    );
    return function (l, j, m) {
      var k = baidu.string(j).toCamelCase(),
        n = m === undefined ? "get" : "set",
        i,
        h;
      k = g[k] || k;
      i = baidu.type(m) === "number" && !~f.indexOf(k) ? m + "px" : m;
      h = (d.hasOwnProperty(k) && d[k][n]) || c[n];
      return h(l, k, i);
    };
  })();
  baidu.dom.extend({
    css: function (a, b) {
      baidu.check("^(?:(?:string(?:,(?:number|string|function))?)|object)$", "baidu.dom.css");
      return baidu._util_.access.call(this, a, b, function (d, c, f) {
        var e = baidu.dom.styleFixer;
        return e ? e(d, c, f) : f === undefined ? this.getCurrentStyle(c) : (d.style[c] = f);
      });
    },
  });
  baidu.dom.extend({
    text: function (f) {
      var e = baidu.dom,
        d = this,
        c = false,
        a;
      if (this.size() <= 0) {
        switch (typeof f) {
          case "undefined":
            return undefined;
            break;
          default:
            return d;
            break;
        }
      }
      var b = function (l) {
        var k,
          h = "",
          j = 0,
          g = l.nodeType;
        if (g) {
          if (g === 1 || g === 9 || g === 11) {
            if (typeof l.textContent === "string") {
              return l.textContent;
            } else {
              for (l = l.firstChild; l; l = l.nextSibling) {
                h += b(l);
              }
            }
          } else {
            if (g === 3 || g === 4) {
              return l.nodeValue;
            }
          }
        } else {
          for (; (k = l[j]); j++) {
            h += b(k);
          }
        }
        return h;
      };
      baidu.forEach(d, function (h, g) {
        var i = e(h);
        if (a) {
          return;
        }
        switch (typeof f) {
          case "undefined":
            a = b(h);
            return a;
            break;
          case "number":
            f = String(f);
          case "string":
            c = true;
            i.empty().append(((h && h.ownerDocument) || document).createTextNode(f));
            break;
          case "function":
            c = true;
            i.text(f.call(h, g, i.text()));
            break;
        }
      });
      return c ? d : a;
    },
  });
  baidu._util_.getWindowOrDocumentWidthOrHeight =
    baidu._util_.getWindowOrDocumentWidthOrHeight ||
    (function () {
      var a = { window: {}, document: {} };
      baidu.forEach(["Width", "Height"], function (d) {
        var c = "client" + d,
          e = "offset" + d,
          b = "scroll" + d;
        a["window"]["get" + d] = function (f) {
          var h = f.document,
            g = h.documentElement[c];
          return (baidu.browser.isStrict && g) || (h.body && h.body[c]) || g;
        };
        a["document"]["get" + d] = function (f) {
          var g = f.documentElement;
          return g[c] >= g[b] ? g[c] : Math.max(f.body[b], g[b], f.body[e], g[e]);
        };
      });
      return function (d, c, b) {
        return a[c][b === "width" ? "getWidth" : "getHeight"](d);
      };
    })();
  baidu.dom.extend({
    width: function (a) {
      return baidu._util_.access.call(this, "width", a, function (f, d, g) {
        var c = g !== undefined,
          b = c && parseFloat(g),
          e = f != null && f == f.window ? "window" : f.nodeType === 9 ? "document" : false;
        if ((c && b < 0) || isNaN(b)) {
          return;
        }
        c && /^(?:\d*\.)?\d+$/.test((g += "")) && (g += "px");
        return e
          ? baidu._util_.getWindowOrDocumentWidthOrHeight(f, e, d)
          : c
            ? (f.style.width = g)
            : baidu._util_.getWidthOrHeight(f, d);
      });
    },
  });
  baidu.dom.extend({
    height: function (a) {
      return baidu._util_.access.call(this, "height", a, function (f, d, g) {
        var c = g !== undefined,
          b = c && parseFloat(g),
          e = f != null && f == f.window ? "window" : f.nodeType === 9 ? "document" : false;
        if ((c && b < 0) || isNaN(b)) {
          return;
        }
        c && /^(?:\d*\.)?\d+$/.test((g += "")) && (g += "px");
        return e
          ? baidu._util_.getWindowOrDocumentWidthOrHeight(f, e, d)
          : c
            ? (f.style.height = g)
            : baidu._util_.getWidthOrHeight(f, d);
      });
    },
  });
  baidu.dom.extend({
    getWindow: function () {
      var a = this.getDocument();
      return this.size() <= 0 ? undefined : a.parentWindow || a.defaultView;
    },
  });
  baidu.dom.extend({
    offset: (function () {
      var a = {
        getDefaultOffset: function (k, i) {
          var c = i.documentElement,
            e = i.body,
            f = i.defaultView,
            g = f ? f.getComputedStyle(k, null) : k.currentStyle,
            h = /^t(?:able|h|d)/i,
            b = k.offsetParent,
            d = k.offsetLeft,
            j = k.offsetTop;
          while ((k = k.parentNode) && k !== e && k !== c) {
            if (baidu.support.fixedPosition && g.position === "fixed") {
              break;
            }
            g = f ? f.getComputedStyle(k, null) : k.currentStyle;
            d -= k.scrollLeft;
            j -= k.scrollTop;
            if (k === b) {
              d += k.offsetLeft;
              j += k.offsetTop;
              if (!baidu.support.hasBorderWidth && !(baidu.support.hasTableCellBorderWidth && h.test(k.nodeName))) {
                d += parseFloat(g.borderLeftWidth) || 0;
                j += parseFloat(g.borderTopWidth) || 0;
              }
              b = k.offsetParent;
            }
          }
          if (~"static,relative".indexOf(g.position)) {
            d += e.offsetLeft;
            j += e.offsetTop;
          }
          if (baidu.support.fixedPosition && g.position === "fixed") {
            d += Math.max(c.scrollLeft, e.scrollLeft);
            j += Math.max(c.scrollTop, e.scrollTop);
          }
          return { left: d, top: j };
        },
        setOffset: function (i, e, d) {
          var h = baidu.dom(i),
            g = baidu.type(e),
            c = h.offset(),
            b = h.getCurrentStyle("left"),
            f = h.getCurrentStyle("top");
          g === "function" && (e = e.call(i, d, c));
          if (!e || (e.left === undefined && e.top === undefined)) {
            return;
          }
          b = parseFloat(b) || 0;
          f = parseFloat(f) || 0;
          e.left != undefined && (i.style.left = e.left - c.left + b + "px");
          e.top != undefined && (i.style.top = e.top - c.top + f + "px");
          h.getCurrentStyle("position") === "static" && (i.style.position = "relative");
        },
        bodyOffset: function (b) {
          var c = baidu.dom(b);
          return {
            left: b.offsetLeft + parseFloat(c.getCurrentStyle("marginLeft")) || 0,
            top: b.offsetTop + parseFloat(c.getCurrentStyle("marginTop")) || 0,
          };
        },
      };
      a.getOffset =
        "getBoundingClientRect" in document.documentElement
          ? function (e, g) {
              if (!e.getBoundingClientRect) {
                return a.getDefaultOffset(e, g);
              }
              var d = e.getBoundingClientRect(),
                f = baidu.dom(g).getWindow(),
                c = g.documentElement,
                b = g.body;
              return {
                left: d.left + (f.pageXOffset || Math.max(c.scrollLeft, b.scrollLeft)) - (c.clientLeft || b.clientLeft),
                top: d.top + (f.pageYOffset || Math.max(c.scrollTop, b.scrollTop)) - (c.clientTop || b.clientTop),
              };
            }
          : a.getDefaultOffset;
      return function (b) {
        if (!b) {
          var e = this[0],
            f = baidu.dom(e).getDocument();
          return a[e === f.body ? "bodyOffset" : "getOffset"](e, f);
        } else {
          baidu.check("^(?:object|function)$", "baidu.dom.offset");
          for (var c = 0, d; (d = this[c]); c++) {
            a.setOffset(d, b, c);
          }
          return this;
        }
      };
    })(),
  });
  baidu.dom.extend({
    removeAttr: function (a) {
      if (arguments.length <= 0 || !a || typeof a !== "string") {
        return this;
      }
      baidu.forEach(this, function (k) {
        var f,
          d,
          b,
          c,
          j,
          e = 0;
        if (k.nodeType === 1) {
          var g = baidu.dom,
            h = baidu._util_;
          d = a.toLowerCase().split(/\s+/);
          c = d.length;
          for (; e < c; e++) {
            b = d[e];
            if (b) {
              f = h.propFix[b] || b;
              j = h.rboolean.test(b);
              if (!j) {
                g(k).attr(b, "");
              }
              k.removeAttribute(baidu.support.getSetAttribute ? b : f);
              if (j && f in k) {
                k[f] = false;
              }
            }
          }
        }
      });
      return this;
    },
  });
  baidu.dom.extend({
    attr: function (b, e) {
      if (arguments.length <= 0 || typeof b === "function") {
        return this;
      }
      var a,
        d = this,
        c = false;
      if (this.size() <= 0) {
        if (b && e) {
          return d;
        } else {
          if (b && !e) {
            return undefined;
          } else {
            return d;
          }
        }
      }
      baidu.forEach(this, function (n, j) {
        if (a) {
          return;
        }
        var k,
          m,
          g,
          i = baidu.dom,
          l = baidu._util_,
          h = n.nodeType;
        if (!n || h === 3 || h === 8 || h === 2) {
          return;
        }
        if (typeof n.getAttribute === "undefined") {
          var o = i(n);
          a = o.prop(b, e);
        }
        switch (typeof b) {
          case "string":
            g = h !== 1 || !baidu._util_.isXML(n);
            if (g) {
              b = b.toLowerCase();
              m = l.attrHooks[b] || (l.rboolean.test(b) ? l.boolHook : l.nodeHook);
            }
            if (typeof e === "undefined") {
              if (m && "get" in m && g && (k = m.get(n, b)) !== null) {
                a = k;
              } else {
                k = n.getAttribute(b);
                a = k === null ? undefined : k;
              }
            } else {
              if (typeof e === "function") {
                c = true;
                var o = i(n);
                o.attr(b, e.call(n, j, o.attr(b)));
              } else {
                c = true;
                var f = { val: true, css: true, html: true, text: true, width: true, height: true, offset: true };
                if (b in f) {
                  a = i(n)[b](e);
                  return;
                }
                if (e === null) {
                  i(n).removeAttr(b);
                  return;
                } else {
                  if (m && "set" in m && g && (k = m.set(n, e, b)) !== undefined) {
                    return k;
                  } else {
                    n.setAttribute(b, "" + e);
                    return e;
                  }
                }
              }
            }
            break;
          case "object":
            c = true;
            var o = i(n);
            for (key in b) {
              o.attr(key, b[key]);
            }
            break;
          default:
            a = d;
            break;
        }
      });
      return c ? this : a;
    },
  });
  baidu.dom.extend({
    before: function () {
      baidu.check(
        "^(?:string|function|HTMLElement|\\$DOM)(?:,(?:string|array|HTMLElement|\\$DOM))*$",
        "baidu.dom.before",
      );
      var a = this[0] && this[0].parentNode,
        c = !a && [],
        b;
      baidu._util_.smartInsert(this, arguments, function (d) {
        a ? a.insertBefore(d, this) : baidu.merge(c, d.childNodes);
      });
      if (c) {
        c = baidu.merge(c, this);
        this.length = 0;
        baidu.merge(this, c);
      }
      return this;
    },
  });
  baidu.dom.extend({
    bind: function (b, c, a) {
      return this.on(b, undefined, c, a);
    },
  });
  baidu.dom.extend({
    children: function (c) {
      var b,
        d = [];
      this.each(function (a) {
        baidu.forEach(this.children || this.childNodes, function (e) {
          e.nodeType == 1 && d.push(e);
        });
      });
      return baidu.dom(baidu.dom.match(d, c));
    },
  });
  baidu.dom.children = function (a) {
    baidu.check("string|HTMLElement", "baidu.dom.children");
    return baidu
      .dom(baidu.isString(a) ? "#" + a : a)
      .children()
      .toArray();
  };
  baidu.dom.extend({
    contents: function () {
      var e = Array.prototype.slice.call(this),
        a = [],
        d;
      for (var b = 0, c; (c = e[b]); b++) {
        d = c.nodeName;
        a.push.apply(
          a,
          baidu.makeArray(
            d && d.toLowerCase() === "iframe" ? c.contentDocument || c.contentWindow.document : c.childNodes,
          ),
        );
      }
      this.length = 0;
      return baidu.merge(this, a);
    },
  });
  baidu.dom._NAME_ATTRS = (function () {
    var a = {
      cellpadding: "cellPadding",
      cellspacing: "cellSpacing",
      colspan: "colSpan",
      rowspan: "rowSpan",
      valign: "vAlign",
      usemap: "useMap",
      frameborder: "frameBorder",
    };
    if (baidu.browser.ie < 8) {
      a["for"] = "htmlFor";
      a["class"] = "className";
    } else {
      a["htmlFor"] = "for";
      a["className"] = "class";
    }
    return a;
  })();
  baidu.dom.extend({
    setAttr: function (b, c) {
      var a = this[0];
      if ("style" == b) {
        a.style.cssText = c;
      } else {
        b = baidu.dom._NAME_ATTRS[b] || b;
        a.setAttribute(b, c);
      }
      return a;
    },
  });
  baidu.dom.create = function (d, a) {
    var e = document.createElement(d),
      b = a || {};
    for (var c in b) {
      baidu.dom.setAttr(e, c, b[c]);
    }
    return e;
  };
  baidu.dom.extend({
    data: (function () {
      var a = baidu.key,
        b = baidu.global("_maps_HTMLElementData");
      return function (c, e) {
        baidu.forEach(this, function (f) {
          !f[a] && (f[a] = baidu.id());
        });
        if (baidu.isString(c)) {
          if (typeof e == "undefined") {
            var d;
            return this[0] && (d = b[this[0][a]]) && d[c];
          }
          baidu.forEach(this, function (g) {
            var f = (b[g[a]] = b[g[a]] || {});
            f[c] = e;
          });
        } else {
          if (baidu.type(c) == "object") {
            baidu.forEach(this, function (g) {
              var f = (b[g[a]] = b[g[a]] || {});
              baidu.forEach(c, function (h) {
                f[h] = c[h];
              });
            });
          }
        }
        return this;
      };
    })(),
  });
  baidu.lang.Class = function () {
    this.guid = baidu.id(this);
  };
  baidu.lang.Class.prototype.dispose = function () {
    baidu.id(this.guid, "delete");
    for (var a in this) {
      typeof this[a] != "function" && delete this[a];
    }
    this.disposed = true;
  };
  baidu.lang.Class.prototype.toString = function () {
    return "[object " + (this._type_ || this.__type || this._className || "Object") + "]";
  };
  baiduInstance = function (a) {
    return baidu.id(a);
  };
  baidu.lang.Class.prototype.un = baidu.lang.Class.prototype.removeEventListener = function (d, c) {
    var b,
      a = this.__listeners;
    if (!a) {
      return;
    }
    if (typeof d == "undefined") {
      for (b in a) {
        delete a[b];
      }
      return;
    }
    d.indexOf("on") && (d = "on" + d);
    if (typeof c == "undefined") {
      delete a[d];
    } else {
      if (a[d]) {
        typeof c == "string" && (c = a[d][c]) && delete a[d][c];
        for (b = a[d].length - 1; b >= 0; b--) {
          if (a[d][b] === c) {
            a[d].splice(b, 1);
          }
        }
      }
    }
  };
  baidu.lang.guid = function () {
    return baidu.id();
  };
  baidu.lang.isString = baidu.isString;
  baidu.lang.Event = function (a, b) {
    this.type = a;
    this.returnValue = true;
    this.target = b || null;
    this.currentTarget = null;
  };
  baidu.lang.Class.prototype.fire = baidu.lang.Class.prototype.dispatchEvent = function (e, a) {
    baidu.lang.isString(e) && (e = new baidu.lang.Event(e));
    !this.__listeners && (this.__listeners = {});
    a = a || {};
    for (var c in a) {
      e[c] = a[c];
    }
    var c,
      g,
      d = this,
      b = d.__listeners,
      f = e.type;
    e.target = e.target || (e.currentTarget = d);
    f.indexOf("on") && (f = "on" + f);
    typeof d[f] == "function" && d[f].apply(d, arguments);
    if (typeof b[f] == "object") {
      for (c = 0, g = b[f].length; c < g; c++) {
        b[f][c] && b[f][c].apply(d, arguments);
      }
    }
    return e.returnValue;
  };
  baidu.lang.Class.prototype.on = baidu.lang.Class.prototype.addEventListener = function (e, d, c) {
    if (typeof d != "function") {
      return;
    }
    !this.__listeners && (this.__listeners = {});
    var b,
      a = this.__listeners;
    e.indexOf("on") && (e = "on" + e);
    typeof a[e] != "object" && (a[e] = []);
    for (b = a[e].length - 1; b >= 0; b--) {
      if (a[e][b] === d) {
        return d;
      }
    }
    a[e].push(d);
    c && typeof c == "string" && (a[e][c] = d);
    return d;
  };
  baidu.lang.createSingle = function (b) {
    var d = new baidu.lang.Class();
    for (var a in b) {
      d[a] = b[a];
    }
    return d;
  };
  baidu.dom.ddManager = baidu.lang.createSingle({ _targetsDroppingOver: {} });
  baidu.dom.extend({
    delegate: function (a, c, d, b) {
      if (typeof d == "function") {
        ((b = d), (d = null));
      }
      return this.on(c, a, d, b);
    },
  });
  baidu.dom.extend({
    filter: function (a) {
      return baidu.dom(baidu.dom.match(this, a));
    },
  });
  baidu.dom.extend({
    remove: function (a, d) {
      arguments.length > 0 && baidu.check("^string(?:,boolean)?$", "baidu.dom.remove");
      var e = a ? this.filter(a) : this;
      for (var b = 0, c; (c = e[b]); b++) {
        if (!d && c.nodeType === 1) {
          baidu._util_.cleanData(c.getElementsByTagName("*"));
          baidu._util_.cleanData([c]);
        }
        c.parentNode && c.parentNode.removeChild(c);
      }
      return this;
    },
  });
  baidu.dom.extend({
    detach: function (a) {
      a && baidu.check("^string$", "baidu.dom.detach");
      return this.remove(a, true);
    },
  });
  baidu.dom._styleFixer = baidu.dom._styleFixer || {};
  baidu.dom._styleFilter = baidu.dom._styleFilter || [];
  baidu.dom._styleFilter.filter = function (b, e, f) {
    for (var a = 0, d = baidu.dom._styleFilter, c; (c = d[a]); a++) {
      if ((c = c[f])) {
        e = c(b, e);
      }
    }
    return e;
  };
  baidu.dom.getStyle = function (c, b) {
    var e = baidu.dom;
    c = e.g(c);
    b = baidu.string.toCamelCase(b);
    var d = c.style[b] || (c.currentStyle ? c.currentStyle[b] : "") || e.getComputedStyle(c, b);
    if (!d || d == "auto") {
      var a = e._styleFixer[b];
      if (a) {
        d = a.get ? a.get(c, b, d) : baidu.dom.getStyle(c, a);
      }
    }
    if ((a = e._styleFilter)) {
      d = a.filter(b, d, "get");
    }
    return d;
  };
  baidu.page = baidu.page || {};
  baidu.page.getScrollTop = function () {
    var a = document;
    return window.pageYOffset || a.documentElement.scrollTop || a.body.scrollTop;
  };
  baidu.page.getScrollLeft = function () {
    var a = document;
    return window.pageXOffset || a.documentElement.scrollLeft || a.body.scrollLeft;
  };
  (function () {
    baidu.page.getMousePosition = function () {
      return { x: baidu.page.getScrollLeft() + a.x, y: baidu.page.getScrollTop() + a.y };
    };
    var a = { x: 0, y: 0 };
    baidu.event.on(document, "onmousemove", function (b) {
      b = window.event || b;
      a.x = b.clientX;
      a.y = b.clientY;
    });
  })();
  baidu.dom.extend({
    off: function (c, a, d) {
      var b = baidu._util_.eventBase,
        e = this;
      if (!c) {
        baidu.forEach(this, function (f) {
          b.removeAll(f);
        });
      } else {
        if (typeof c == "string") {
          if (typeof a == "function") {
            ((d = a), (a = null));
          }
          c = c.split(/[ ,]/);
          baidu.forEach(this, function (f) {
            baidu.forEach(c, function (g) {
              b.remove(f, g, d, a);
            });
          });
        } else {
          if (typeof c == "object") {
            baidu.forEach(c, function (f, g) {
              e.off(g, a, f);
            });
          }
        }
      }
      return this;
    },
  });
  baidu.event.un = baidu.un = function (a, c, b) {
    a = baidu.dom.g(a);
    baidu.dom(a).off(c.replace(/^\s*on/, ""), b);
    return a;
  };
  baidu.event.preventDefault = function (a) {
    a.originalEvent && (a = a.originalEvent);
    if (a.preventDefault) {
      a.preventDefault();
    } else {
      a.returnValue = false;
    }
  };
  (function () {
    var m = false,
      i,
      h,
      e,
      d,
      c,
      f,
      l,
      a,
      k,
      n;
    baidu.dom.drag = function (p, o) {
      if (!(i = baidu.dom.g(p))) {
        return false;
      }
      h = baidu.object.extend({ autoStop: true, capture: true, interval: 16 }, o);
      a = f = parseInt(baidu.dom.getStyle(i, "left")) || 0;
      k = l = parseInt(baidu.dom.getStyle(i, "top")) || 0;
      m = true;
      setTimeout(function () {
        var r = baidu.page.getMousePosition();
        e = h.mouseEvent ? baidu.page.getScrollLeft() + h.mouseEvent.clientX : r.x;
        d = h.mouseEvent ? baidu.page.getScrollTop() + h.mouseEvent.clientY : r.y;
        clearInterval(c);
        c = setInterval(b, h.interval);
      }, 1);
      var q = baidu.dom(document);
      h.autoStop && q.on("mouseup", j);
      q.on("selectstart", g);
      if (h.capture && i.setCapture) {
        i.setCapture();
      } else {
        if (h.capture && window.captureEvents) {
          window.captureEvents(Event.MOUSEMOVE | Event.MOUSEUP);
        }
      }
      n = document.body.style.MozUserSelect;
      document.body.style.MozUserSelect = "none";
      baidu.isFunction(h.ondragstart) && h.ondragstart(i, h);
      return {
        stop: j,
        dispose: j,
        update: function (r) {
          baidu.object.extend(h, r);
        },
      };
    };
    function j() {
      m = false;
      clearInterval(c);
      if (h.capture && i.releaseCapture) {
        i.releaseCapture();
      } else {
        if (h.capture && window.captureEvents) {
          window.captureEvents(Event.MOUSEMOVE | Event.MOUSEUP);
        }
      }
      document.body.style.MozUserSelect = n;
      var o = baidu.dom(document);
      o.off("selectstart", g);
      h.autoStop && o.off("mouseup", j);
      baidu.isFunction(h.ondragend) && h.ondragend(i, h, { left: a, top: k });
    }
    function b(s) {
      if (!m) {
        clearInterval(c);
        return;
      }
      var p = h.range || [],
        o = baidu.page.getMousePosition(),
        q = f + o.x - e,
        r = l + o.y - d;
      if (baidu.isObject(p) && p.length == 4) {
        q = Math.max(p[3], q);
        q = Math.min(p[1] - i.offsetWidth, q);
        r = Math.max(p[0], r);
        r = Math.min(p[2] - i.offsetHeight, r);
      }
      i.style.left = q + "px";
      i.style.top = r + "px";
      a = q;
      k = r;
      baidu.isFunction(h.ondrag) && h.ondrag(i, h, { left: a, top: k });
    }
    function g(o) {
      return baidu.event.preventDefault(o, false);
    }
  })();
  baidu.dom.extend({
    setStyle: function (b, c) {
      var d = baidu.dom,
        a;
      element = this[0];
      b = baidu.string.toCamelCase(b);
      if ((a = d._styleFilter)) {
        c = a.filter(b, c, "set");
      }
      a = d._styleFixer[b];
      a && a.set ? a.set(element, c, b) : (element.style[a || b] = c);
      return element;
    },
  });
  baidu.dom.draggable = function (b, k) {
    k = baidu.object.extend(
      {
        toggle: function () {
          return true;
        },
      },
      k,
    );
    k.autoStop = true;
    b = baidu.dom.g(b);
    k.handler = k.handler || b;
    var a,
      h = ["ondragstart", "ondrag", "ondragend"],
      c = h.length - 1,
      d,
      j,
      f = {
        dispose: function () {
          j && j.stop();
          baidu.event.un(k.handler, "onmousedown", g);
          baidu.lang.Class.prototype.dispose.call(f);
        },
      },
      e = this;
    if ((a = baidu.dom.ddManager)) {
      for (; c >= 0; c--) {
        d = h[c];
        k[d] = (function (i) {
          var l = k[i];
          return function () {
            baidu.lang.isFunction(l) && l.apply(e, arguments);
            a.dispatchEvent(i, { DOM: b });
          };
        })(d);
      }
    }
    if (b) {
      function g(l) {
        var i = (k.mouseEvent = window.event || l);
        k.mouseEvent = { clientX: i.clientX, clientY: i.clientY };
        if (i.button > 1 || (baidu.lang.isFunction(k.toggle) && !k.toggle())) {
          return;
        }
        if (baidu.lang.isFunction(k.onbeforedragstart)) {
          k.onbeforedragstart(b);
        }
        j = baidu.dom.drag(b, k);
        f.stop = j.stop;
        f.update = j.update;
        baidu.event.preventDefault(i);
      }
      baidu.event.on(k.handler, "onmousedown", g);
    }
    return {
      cancel: function () {
        f.dispose();
      },
    };
  };
  baidu.dom.getPosition = function (a) {
    a = baidu.dom.g(a);
    var j = baidu.dom.getDocument(a),
      d = baidu.browser,
      g = baidu.dom.getStyle,
      c =
        d.isGecko > 0 &&
        j.getBoxObjectFor &&
        g(a, "position") == "absolute" &&
        (a.style.top === "" || a.style.left === ""),
      h = { left: 0, top: 0 },
      f = d.ie && !d.isStrict ? j.body : j.documentElement,
      k,
      b;
    if (a == f) {
      return h;
    }
    if (a.getBoundingClientRect) {
      b = a.getBoundingClientRect();
      h.left = Math.floor(b.left) + Math.max(j.documentElement.scrollLeft, j.body.scrollLeft);
      h.top = Math.floor(b.top) + Math.max(j.documentElement.scrollTop, j.body.scrollTop);
      h.left -= j.documentElement.clientLeft;
      h.top -= j.documentElement.clientTop;
      var i = j.body,
        l = parseInt(g(i, "borderLeftWidth")),
        e = parseInt(g(i, "borderTopWidth"));
      if (d.ie && !d.isStrict) {
        h.left -= isNaN(l) ? 2 : l;
        h.top -= isNaN(e) ? 2 : e;
      }
    } else {
      k = a;
      do {
        h.left += k.offsetLeft;
        h.top += k.offsetTop;
        if (d.isWebkit > 0 && g(k, "position") == "fixed") {
          h.left += j.body.scrollLeft;
          h.top += j.body.scrollTop;
          break;
        }
        k = k.offsetParent;
      } while (k && k != a);
      if (d.opera > 0 || (d.isWebkit > 0 && g(a, "position") == "absolute")) {
        h.top -= j.body.offsetTop;
      }
      k = a.offsetParent;
      while (k && k != j.body) {
        h.left -= k.scrollLeft;
        if (!d.opera || k.tagName != "TR") {
          h.top -= k.scrollTop;
        }
        k = k.offsetParent;
      }
    }
    return h;
  };
  baidu.dom.intersect = function (i, h) {
    var f = baidu.dom.g,
      e = baidu.dom.getPosition,
      a = Math.max,
      c = Math.min;
    i = f(i);
    h = f(h);
    var d = e(i),
      b = e(h);
    return (
      a(d.left, b.left) <= c(d.left + i.offsetWidth, b.left + h.offsetWidth) &&
      a(d.top, b.top) <= c(d.top + i.offsetHeight, b.top + h.offsetHeight)
    );
  };
  baidu.dom.droppable = function (e, c) {
    c = c || {};
    var d = baidu.dom.ddManager,
      g = baidu.dom.g(e),
      b = baidu.lang.guid(),
      f = function (j) {
        var i = d._targetsDroppingOver,
          h = { trigger: j.DOM, reciever: g };
        if (baidu.dom.intersect(g, j.DOM)) {
          if (!i[b]) {
            typeof c.ondropover == "function" && c.ondropover.call(g, h);
            d.dispatchEvent("ondropover", h);
            i[b] = true;
          }
        } else {
          if (i[b]) {
            typeof c.ondropout == "function" && c.ondropout.call(g, h);
            d.dispatchEvent("ondropout", h);
          }
          delete i[b];
        }
      },
      a = function (i) {
        var h = { trigger: i.DOM, reciever: g };
        if (baidu.dom.intersect(g, i.DOM)) {
          typeof c.ondrop == "function" && c.ondrop.call(g, h);
          d.dispatchEvent("ondrop", h);
        }
        delete d._targetsDroppingOver[b];
      };
    d.addEventListener("ondrag", f);
    d.addEventListener("ondragend", a);
    return {
      cancel: function () {
        d.removeEventListener("ondrag", f);
        d.removeEventListener("ondragend", a);
      },
    };
  };
  baidu.dom.extend({
    eq: function (a) {
      baidu.check("number", "baidu.dom.eq");
      return baidu.dom(this.get(a));
    },
  });
  baidu.dom.extend({
    find: function (b) {
      var c = [],
        d,
        f = "__tangram__find__",
        e = baidu.dom();
      switch (baidu.type(b)) {
        case "string":
          this.each(function () {
            baidu.merge(e, baidu.query(b, this));
          });
          break;
        case "HTMLElement":
          d = b.tagName + "#" + (b.id ? b.id : (b.id = f));
          this.each(function () {
            if (baidu.query(d, this).length > 0) {
              c.push(b);
            }
          });
          b.id == f && (b.id = "");
          if (c.length > 0) {
            baidu.merge(e, c);
          }
          break;
        case "$DOM":
          c = b.get();
          this.each(function () {
            baidu.forEach(baidu.query("*", this), function (g) {
              for (var a = 0, h = c.length; a < h; a++) {
                g === c[a] && (e[e.length++] = c[a]);
              }
            });
          });
          break;
      }
      return e;
    },
  });
  baidu.dom.extend({
    first: function () {
      return baidu.dom(this[0]);
    },
  });
  baidu.dom.first = function (a) {
    baidu.isString(a) && (a = "#" + a);
    return baidu.dom(a).children()[0];
  };
  baidu.dom.getAncestorBy = function (a, b) {
    a = baidu.dom.g(a);
    while ((a = a.parentNode) && a.nodeType == 1) {
      if (b(a)) {
        return a;
      }
    }
    return null;
  };
  baidu.dom.getAncestorByClass = function (a, b) {
    a = baidu.dom.g(a);
    b = new RegExp("(^|\\s)" + baidu.string.trim(b) + "(\\s|\x24)");
    while ((a = a.parentNode) && a.nodeType == 1) {
      if (b.test(a.className)) {
        return a;
      }
    }
    return null;
  };
  baidu.dom.getAncestorByTag = function (b, a) {
    b = baidu.dom.g(b);
    a = a.toUpperCase();
    while ((b = b.parentNode) && b.nodeType == 1) {
      if (b.tagName == a) {
        return b;
      }
    }
    return null;
  };
  baidu.dom.extend({
    getAttr: function (a) {
      element = this[0];
      if ("style" == a) {
        return element.style.cssText;
      }
      a = baidu.dom._NAME_ATTRS[a] || a;
      return element.getAttribute(a);
    },
  });
  baidu.dom.getParent = function (a) {
    a = baidu.dom._g(a);
    return a.parentElement || a.parentNode || null;
  };
  baidu.dom.extend({
    getText: function () {
      var b = "",
        d,
        c = 0,
        a;
      element = this[0];
      if (element.nodeType === 3 || element.nodeType === 4) {
        b += element.nodeValue;
      } else {
        if (element.nodeType !== 8) {
          d = element.childNodes;
          for (a = d.length; c < a; c++) {
            b += baidu.dom.getText(d[c]);
          }
        }
      }
      return b;
    },
  });
  baidu.dom.extend({
    has: function (b) {
      var c = [],
        d = baidu.dom(document.body);
      baidu.forEach(this, function (a) {
        d[0] = a;
        d.find(b).length && c.push(a);
      });
      return baidu.dom(c);
    },
  });
  baidu.dom.extend({
    hasAttr: function (b) {
      element = this[0];
      var a = element.attributes.getNamedItem(b);
      return !!(a && a.specified);
    },
  });
  baidu.dom.extend({
    hasClass: function (c) {
      if (arguments.length <= 0 || typeof c === "function") {
        return this;
      }
      if (this.size() <= 0) {
        return false;
      }
      c = c.replace(/^\s+/g, "").replace(/\s+$/g, "").replace(/\s+/g, " ");
      var b = c.split(" ");
      var a;
      baidu.forEach(this, function (e) {
        var f = e.className;
        for (var d = 0; d < b.length; d++) {
          if ((" " + f + " ").indexOf(" " + b[d] + " ") == -1) {
            a = false;
            return;
          }
        }
        if (a !== false) {
          a = true;
          return;
        }
      });
      return a;
    },
  });
  (function () {
    var f,
      c,
      a,
      h = /^margin/,
      e = /^-?(?:\d*\.)?\d+(?!px)[^\d\s]+$/i,
      g = /^(top|right|bottom|left)$/,
      b = {};
    baidu.extend(baidu._util_, {
      showHide: function (n, i) {
        var m,
          o,
          j = [],
          k = 0,
          l = n.length;
        for (; k < l; k++) {
          m = n[k];
          if (!m.style) {
            continue;
          }
          j[k] = baidu._util_._data(m, "olddisplay");
          if (i) {
            if (!j[k] && m.style.display === "none") {
              m.style.display = "";
            }
            if (
              (m.style.display === "" && a(m, "display") === "none") ||
              !baidu.dom.contains(m.ownerDocument.documentElement, m)
            ) {
              j[k] = baidu._util_._data(m, "olddisplay", d(m.nodeName));
            }
          } else {
            o = a(m, "display");
            if (!j[k] && o !== "none") {
              baidu._util_._data(m, "olddisplay", o);
            }
          }
        }
        for (k = 0; k < l; k++) {
          m = n[k];
          if (!m.style) {
            continue;
          }
          if (!i || m.style.display === "none" || m.style.display === "") {
            m.style.display = i ? j[k] || "" : "none";
          }
        }
        return n;
      },
    });
    if (window.getComputedStyle) {
      a = function (n, j) {
        var i,
          l,
          m = getComputedStyle(n, null),
          k = n.style;
        if (m) {
          i = m[j];
          if (i === "" && !baidu.dom.contains(n.ownerDocument.documentElement, n)) {
            i = baidu.dom(n).css(j);
          }
          if (!baidu.support.pixelMargin && h.test(j) && e.test(i)) {
            l = k.width;
            k.width = i;
            i = m.width;
            k.width = l;
          }
        }
        return i;
      };
    } else {
      if (document.documentElement.currentStyle) {
        a = function (m, k) {
          var n,
            i,
            j = m.currentStyle && m.currentStyle[k],
            l = m.style;
          if (j == null && l && l[k]) {
            j = l[k];
          }
          if (e.test(j) && !g.test(k)) {
            n = l.left;
            i = m.runtimeStyle && m.runtimeStyle.left;
            if (i) {
              m.runtimeStyle.left = m.currentStyle.left;
            }
            l.left = k === "fontSize" ? "1em" : j;
            j = l.pixelLeft + "px";
            l.left = n;
            if (i) {
              m.runtimeStyle.left = i;
            }
          }
          return j === "" ? "auto" : j;
        };
      }
    }
    function d(k) {
      if (b[k]) {
        return b[k];
      }
      var i = baidu.dom("<" + k + ">").appendTo(document.body),
        j = i.css("display");
      i.remove();
      if (j === "none" || j === "") {
        c = document.body.appendChild(
          c || baidu.extend(document.createElement("iframe"), { frameBorder: 0, width: 0, height: 0 }),
        );
        if (!f || !c.createElement) {
          f = (c.contentWindow || c.contentDocument).document;
          f.write("<!doctype html><html><body>");
          f.close();
        }
        i = f.body.appendChild(f.createElement(k));
        j = a(i, "display");
        document.body.removeChild(c);
      }
      b[k] = j;
      return j;
    }
  })();
  baidu.dom.extend({
    hide: function () {
      baidu._util_.showHide(this);
      return this;
    },
  });
  baidu.dom.extend({
    innerHeight: function () {
      if (this.size() <= 0) {
        return 0;
      }
      var b = this[0],
        a = b != null && b === b.window ? "window" : b.nodeType === 9 ? "document" : false;
      return a
        ? baidu._util_.getWindowOrDocumentWidthOrHeight(b, a, "height")
        : baidu._util_.getWidthOrHeight(b, "height", "padding");
    },
  });
  baidu.dom.extend({
    innerWidth: function () {
      if (this.size() <= 0) {
        return 0;
      }
      var b = this[0],
        a = b != null && b === b.window ? "window" : b.nodeType === 9 ? "document" : false;
      return a
        ? baidu._util_.getWindowOrDocumentWidthOrHeight(b, a, "width")
        : baidu._util_.getWidthOrHeight(b, "width", "padding");
    },
  });
  baidu.dom.extend({
    insertAfter: function (a) {
      baidu.check("^(?:string|HTMLElement|\\$DOM)$", "baidu.dom.insertAfter");
      baidu._util_.smartInsertTo(
        this,
        a,
        function (b) {
          this.parentNode.insertBefore(b, this.nextSibling);
        },
        "after",
      );
      return this;
    },
  });
  baidu.dom.insertAfter = function (c, b) {
    var a = baidu.dom._g;
    return baidu.dom(a(c)).insertAfter(a(b))[0];
  };
  baidu.dom.extend({
    insertBefore: function (a) {
      baidu.check("^(?:string|HTMLElement|\\$DOM)$", "baidu.dom.insertBefore");
      baidu._util_.smartInsertTo(
        this,
        a,
        function (b) {
          this.parentNode.insertBefore(b, this);
        },
        "before",
      );
      return this;
    },
  });
  baidu.dom.insertBefore = function (c, b) {
    var a = baidu.dom._g;
    return baidu.dom(a(c)).insertBefore(a(b))[0];
  };
  baidu.dom.extend({
    insertHTML: function (a, c) {
      element = this[0];
      var b, d;
      if (element.insertAdjacentHTML && !baidu.browser.opera) {
        element.insertAdjacentHTML(a, c);
      } else {
        b = element.ownerDocument.createRange();
        a = a.toUpperCase();
        if (a == "AFTERBEGIN" || a == "BEFOREEND") {
          b.selectNodeContents(element);
          b.collapse(a == "AFTERBEGIN");
        } else {
          d = a == "BEFOREBEGIN";
          b[d ? "setStartBefore" : "setEndAfter"](element);
          b.collapse(d);
        }
        b.insertNode(b.createContextualFragment(c));
      }
      return element;
    },
  });
  baidu.dom.extend({
    last: function () {
      return baidu.dom(this.get(-1));
    },
  });
  baidu.dom.last = function (a) {
    a = baidu.dom.g(a);
    for (var b = a.lastChild; b; b = b.previousSibling) {
      if (b.nodeType == 1) {
        return b;
      }
    }
    return null;
  };
  baidu.dom.extend({
    next: function (a) {
      var b = baidu.dom();
      baidu.forEach(this, function (c) {
        while ((c = c.nextSibling) && c && c.nodeType != 1) {}
        c && (b[b.length++] = c);
      });
      return a ? b.filter(a) : b;
    },
  });
  baidu.dom.extend({
    nextAll: function (a) {
      var b = [];
      baidu.forEach(this, function (c) {
        while ((c = c.nextSibling)) {
          c && c.nodeType == 1 && b.push(c);
        }
      });
      return baidu.dom(baidu.dom.match(b, a));
    },
  });
  baidu.dom.extend({
    nextUntil: function (a, b) {
      var c = baidu.array();
      baidu.forEach(this, function (f) {
        var e = baidu.array();
        while ((f = f.nextSibling)) {
          f && f.nodeType == 1 && e.push(f);
        }
        if (a && e.length) {
          var d = baidu.dom.match(e, a);
          if (d.length) {
            e = e.slice(0, e.indexOf(d[0]));
          }
        }
        baidu.merge(c, e);
      });
      return baidu.dom(baidu.dom.match(c, b));
    },
  });
  baidu.dom.extend({
    not: function (b) {
      var e,
        d,
        g,
        f = this.get(),
        c = baidu.isArray(b) ? b : baidu.dom.match(this, b);
      for (e = f.length - 1; e > -1; e--) {
        for (d = 0, g = c.length; d < g; d++) {
          c[d] === f[e] && f.splice(e, 1);
        }
      }
      return baidu.dom(f);
    },
  });
  baidu.dom.extend({
    offsetParent: function () {
      return this.map(function () {
        var b = this.offsetParent || document.body,
          a = /^(?:body|html)$/i;
        while (b && baidu.dom(b).getCurrentStyle("position") === "static" && !a.test(b.nodeName)) {
          b = b.offsetParent;
        }
        return b;
      });
    },
  });
  baidu.dom.extend({
    one: function (b, e, a) {
      var c = this;
      if (typeof e == "function") {
        ((a = e), (e = undefined));
      }
      if (typeof b == "object" && b) {
        baidu.forEach(
          b,
          function (g, f) {
            this.one(f, e, g);
          },
          this,
        );
        return this;
      }
      var d = function () {
        baidu.dom(this).off(b, d);
        return a.apply(c, arguments);
      };
      this.each(function () {
        var f = baidu.id(this);
        a["_" + f + "_" + b] = d;
      });
      return this.on(b, e, d);
    },
  });
  baidu.dom.opacity = function (b, a) {
    b = baidu.dom.g(b);
    if (!baidu.browser.ie) {
      b.style.opacity = a;
      b.style.KHTMLOpacity = a;
    } else {
      b.style.filter = "progid:DXImageTransform.Microsoft.Alpha(opacity:" + Math.floor(a * 100) + ")";
    }
  };
  baidu.dom.extend({
    outerHeight: function (c) {
      if (this.size() <= 0) {
        return 0;
      }
      var b = this[0],
        a = b != null && b === b.window ? "window" : b.nodeType === 9 ? "document" : false;
      return a
        ? baidu._util_.getWindowOrDocumentWidthOrHeight(b, a, "height")
        : baidu._util_.getWidthOrHeight(b, "height", "padding|border" + (c ? "|margin" : ""));
    },
  });
  baidu.dom.extend({
    outerWidth: function (c) {
      if (this.size() <= 0) {
        return 0;
      }
      var b = this[0],
        a = b != null && b === b.window ? "window" : b.nodeType === 9 ? "document" : false;
      return a
        ? baidu._util_.getWindowOrDocumentWidthOrHeight(b, a, "width")
        : baidu._util_.getWidthOrHeight(b, "width", "padding|border" + (c ? "|margin" : ""));
    },
  });
  baidu.dom.extend({
    parent: function (a) {
      var b = [];
      baidu.forEach(this, function (c) {
        (c = c.parentNode) && c.nodeType == 1 && b.push(c);
      });
      return baidu.dom(baidu.dom.match(b, a));
    },
  });
  baidu.dom.extend({
    parents: function (a) {
      var b = [];
      baidu.forEach(this, function (d) {
        var c = [];
        while ((d = d.parentNode) && d.nodeType == 1) {
          c.push(d);
        }
        baidu.merge(b, c);
      });
      return baidu.dom(baidu.dom.match(b, a));
    },
  });
  baidu.dom.extend({
    parentsUntil: function (a, b) {
      baidu.check("(string|HTMLElement)(,.+)?", "baidu.dom.parentsUntil");
      var c = [];
      baidu.forEach(this, function (f) {
        var e = baidu.array();
        while ((f = f.parentNode) && f.nodeType == 1) {
          e.push(f);
        }
        if (a && e.length) {
          var d = baidu.dom.match(e, a);
          if (d.length) {
            e = e.slice(0, e.indexOf(d[0]));
          }
        }
        baidu.merge(c, e);
      });
      return baidu.dom(baidu.dom.match(c, b));
    },
  });
  baidu.dom.extend({
    position: function () {
      if (this.size() <= 0) {
        return 0;
      }
      var a = /^(?:body|html)$/i,
        d = this.offset(),
        b = this.offsetParent(),
        c = a.test(b[0].nodeName) ? { left: 0, top: 0 } : b.offset();
      d.left -= parseFloat(this.getCurrentStyle("marginLeft")) || 0;
      d.top -= parseFloat(this.getCurrentStyle("marginTop")) || 0;
      c.left += parseFloat(b.getCurrentStyle("borderLeftWidth")) || 0;
      c.top += parseFloat(b.getCurrentStyle("borderTopWidth")) || 0;
      return { left: d.left - c.left, top: d.top - c.top };
    },
  });
  baidu.dom.extend({
    prepend: function () {
      baidu.check(
        "^(?:string|function|HTMLElement|\\$DOM)(?:,(?:string|array|HTMLElement|\\$DOM))*$",
        "baidu.dom.prepend",
      );
      baidu._util_.smartInsert(this, arguments, function (a) {
        this.nodeType === 1 && this.insertBefore(a, this.firstChild);
      });
      return this;
    },
  });
  baidu.dom.extend({
    prependTo: function (a) {
      baidu.check("^(?:string|HTMLElement|\\$DOM)$", "baidu.dom.prependTo");
      baidu._util_.smartInsertTo(this, a, function (b) {
        this.insertBefore(b, this.firstChild);
      });
      return this;
    },
  });
  baidu.dom.extend({
    prev: function (a) {
      var b = [];
      baidu.forEach(this, function (c) {
        while ((c = c.previousSibling)) {
          if (c.nodeType == 1) {
            b.push(c);
            break;
          }
        }
      });
      return baidu.dom(baidu.dom.match(b, a));
    },
  });
  baidu.dom.extend({
    prevAll: function (a) {
      var b = baidu.array();
      baidu.forEach(this, function (d) {
        var c = [];
        while ((d = d.previousSibling)) {
          d.nodeType == 1 && c.push(d);
        }
        baidu.merge(b, c.reverse());
      });
      return baidu.dom(typeof a == "string" ? baidu.dom.match(b, a) : b.unique());
    },
  });
  baidu.dom.extend({
    prevUntil: function (a, b) {
      baidu.check("(string|HTMLElement)(,.+)?", "baidu.dom.prevUntil");
      var c = [];
      baidu.forEach(this, function (f) {
        var e = baidu.array();
        while ((f = f.previousSibling)) {
          f && f.nodeType == 1 && e.push(f);
        }
        if (a && e.length) {
          var d = baidu.dom.match(e, a);
          if (d.length) {
            e = e.slice(0, e.indexOf(d[0]));
          }
        }
        baidu.merge(c, e);
      });
      return baidu.dom(baidu.dom.match(c, b));
    },
  });
  baidu.string.extend({
    escapeReg: function () {
      return this.replace(new RegExp("([.*+?^=!:\x24{}()|[\\]/\\\\])", "g"), "\\\x241");
    },
  });
  baidu.dom.q = function (h, e, b) {
    var j = [],
      d = baidu.string.trim,
      g,
      f,
      a,
      c;
    if (!(h = d(h))) {
      return j;
    }
    if ("undefined" == typeof e) {
      e = document;
    } else {
      e = baidu.dom.g(e);
      if (!e) {
        return j;
      }
    }
    b && (b = d(b).toUpperCase());
    if (e.getElementsByClassName) {
      a = e.getElementsByClassName(h);
      g = a.length;
      for (f = 0; f < g; f++) {
        c = a[f];
        if (b && c.tagName != b) {
          continue;
        }
        j[j.length] = c;
      }
    } else {
      h = new RegExp("(^|\\s)" + baidu.string.escapeReg(h) + "(\\s|\x24)");
      a = b ? e.getElementsByTagName(b) : e.all || e.getElementsByTagName("*");
      g = a.length;
      for (f = 0; f < g; f++) {
        c = a[f];
        h.test(c.className) && (j[j.length] = c);
      }
    }
    return j;
  };
  (function () {
    var n =
        /((?:\((?:\([^()]+\)|[^()]+)+\)|\[(?:\[[^\[\]]*\]|['"][^'"]*['"]|[^\[\]'"]+)+\]|\\.|[^ >+~,(\[\\]+)+|[>+~])(\s*,\s*)?((?:.|\r|\n)*)/g,
      i = "sizcache" + (Math.random() + "").replace(".", ""),
      o = 0,
      r = Object.prototype.toString,
      h = false,
      g = true,
      q = /\\/g,
      u = /\r\n/g,
      w = /\W/;
    [0, 0].sort(function () {
      g = false;
      return 0;
    });
    var d = function (B, e, E, F) {
      E = E || [];
      e = e || document;
      var H = e;
      if (e.nodeType !== 1 && e.nodeType !== 9) {
        return [];
      }
      if (!B || typeof B !== "string") {
        return E;
      }
      var y,
        J,
        M,
        x,
        I,
        L,
        K,
        D,
        A = true,
        z = d.isXML(e),
        C = [],
        G = B;
      do {
        n.exec("");
        y = n.exec(G);
        if (y) {
          G = y[3];
          C.push(y[1]);
          if (y[2]) {
            x = y[3];
            break;
          }
        }
      } while (y);
      if (C.length > 1 && j.exec(B)) {
        if (C.length === 2 && k.relative[C[0]]) {
          J = s(C[0] + C[1], e, F);
        } else {
          J = k.relative[C[0]] ? [e] : d(C.shift(), e);
          while (C.length) {
            B = C.shift();
            if (k.relative[B]) {
              B += C.shift();
            }
            J = s(B, J, F);
          }
        }
      } else {
        if (
          !F &&
          C.length > 1 &&
          e.nodeType === 9 &&
          !z &&
          k.match.ID.test(C[0]) &&
          !k.match.ID.test(C[C.length - 1])
        ) {
          I = d.find(C.shift(), e, z);
          e = I.expr ? d.filter(I.expr, I.set)[0] : I.set[0];
        }
        if (e) {
          I = F
            ? { expr: C.pop(), set: l(F) }
            : d.find(C.pop(), C.length === 1 && (C[0] === "~" || C[0] === "+") && e.parentNode ? e.parentNode : e, z);
          J = I.expr ? d.filter(I.expr, I.set) : I.set;
          if (C.length > 0) {
            M = l(J);
          } else {
            A = false;
          }
          while (C.length) {
            L = C.pop();
            K = L;
            if (!k.relative[L]) {
              L = "";
            } else {
              K = C.pop();
            }
            if (K == null) {
              K = e;
            }
            k.relative[L](M, K, z);
          }
        } else {
          M = C = [];
        }
      }
      if (!M) {
        M = J;
      }
      if (!M) {
        d.error(L || B);
      }
      if (r.call(M) === "[object Array]") {
        if (!A) {
          E.push.apply(E, M);
        } else {
          if (e && e.nodeType === 1) {
            for (D = 0; M[D] != null; D++) {
              if (M[D] && (M[D] === true || (M[D].nodeType === 1 && d.contains(e, M[D])))) {
                E.push(J[D]);
              }
            }
          } else {
            for (D = 0; M[D] != null; D++) {
              if (M[D] && M[D].nodeType === 1) {
                E.push(J[D]);
              }
            }
          }
        }
      } else {
        l(M, E);
      }
      if (x) {
        d(x, H, E, F);
        d.uniqueSort(E);
      }
      return E;
    };
    d.uniqueSort = function (x) {
      if (p) {
        h = g;
        x.sort(p);
        if (h) {
          for (var e = 1; e < x.length; e++) {
            if (x[e] === x[e - 1]) {
              x.splice(e--, 1);
            }
          }
        }
      }
      return x;
    };
    d.matches = function (e, x) {
      return d(e, null, null, x);
    };
    d.matchesSelector = function (e, x) {
      return d(x, null, null, [e]).length > 0;
    };
    d.find = function (D, e, E) {
      var C, y, A, z, B, x;
      if (!D) {
        return [];
      }
      for (y = 0, A = k.order.length; y < A; y++) {
        B = k.order[y];
        if ((z = k.leftMatch[B].exec(D))) {
          x = z[1];
          z.splice(1, 1);
          if (x.substr(x.length - 1) !== "\\") {
            z[1] = (z[1] || "").replace(q, "");
            C = k.find[B](z, e, E);
            if (C != null) {
              D = D.replace(k.match[B], "");
              break;
            }
          }
        }
      }
      if (!C) {
        C = typeof e.getElementsByTagName !== "undefined" ? e.getElementsByTagName("*") : [];
      }
      return { set: C, expr: D };
    };
    d.filter = function (H, G, K, A) {
      var C,
        e,
        F,
        M,
        J,
        x,
        z,
        B,
        I,
        y = H,
        L = [],
        E = G,
        D = G && G[0] && d.isXML(G[0]);
      while (H && G.length) {
        for (F in k.filter) {
          if ((C = k.leftMatch[F].exec(H)) != null && C[2]) {
            x = k.filter[F];
            z = C[1];
            e = false;
            C.splice(1, 1);
            if (z.substr(z.length - 1) === "\\") {
              continue;
            }
            if (E === L) {
              L = [];
            }
            if (k.preFilter[F]) {
              C = k.preFilter[F](C, E, K, L, A, D);
              if (!C) {
                e = M = true;
              } else {
                if (C === true) {
                  continue;
                }
              }
            }
            if (C) {
              for (B = 0; (J = E[B]) != null; B++) {
                if (J) {
                  M = x(J, C, B, E);
                  I = A ^ M;
                  if (K && M != null) {
                    if (I) {
                      e = true;
                    } else {
                      E[B] = false;
                    }
                  } else {
                    if (I) {
                      L.push(J);
                      e = true;
                    }
                  }
                }
              }
            }
            if (M !== undefined) {
              if (!K) {
                E = L;
              }
              H = H.replace(k.match[F], "");
              if (!e) {
                return [];
              }
              break;
            }
          }
        }
        if (H === y) {
          if (e == null) {
            d.error(H);
          } else {
            break;
          }
        }
        y = H;
      }
      return E;
    };
    d.error = function (e) {
      throw "Syntax error, unrecognized expression: " + e;
    };
    var b = (d.getText = function (A) {
      var y,
        z,
        e = A.nodeType,
        x = "";
      if (e) {
        if (e === 1) {
          if (typeof A.textContent === "string") {
            return A.textContent;
          } else {
            if (typeof A.innerText === "string") {
              return A.innerText.replace(u, "");
            } else {
              for (A = A.firstChild; A; A = A.nextSibling) {
                x += b(A);
              }
            }
          }
        } else {
          if (e === 3 || e === 4) {
            return A.nodeValue;
          }
        }
      } else {
        for (y = 0; (z = A[y]); y++) {
          if (z.nodeType !== 8) {
            x += b(z);
          }
        }
      }
      return x;
    });
    var k = (d.selectors = {
      order: ["ID", "NAME", "TAG"],
      match: {
        ID: /#((?:[\w\u00c0-\uFFFF\-]|\\.)+)/,
        CLASS: /\.((?:[\w\u00c0-\uFFFF\-]|\\.)+)/,
        NAME: /\[name=['"]*((?:[\w\u00c0-\uFFFF\-]|\\.)+)['"]*\]/,
        ATTR: /\[\s*((?:[\w\u00c0-\uFFFF\-]|\\.)+)\s*(?:(\S?=)\s*(?:(['"])(.*?)\3|(#?(?:[\w\u00c0-\uFFFF\-]|\\.)*)|)|)\s*\]/,
        TAG: /^((?:[\w\u00c0-\uFFFF\*\-]|\\.)+)/,
        CHILD: /:(only|nth|last|first)-child(?:\(\s*(even|odd|(?:[+\-]?\d+|(?:[+\-]?\d*)?n\s*(?:[+\-]\s*\d+)?))\s*\))?/,
        POS: /:(nth|eq|gt|lt|first|last|even|odd)(?:\((\d*)\))?(?=[^\-]|$)/,
        PSEUDO: /:((?:[\w\u00c0-\uFFFF\-]|\\.)+)(?:\((['"]?)((?:\([^\)]+\)|[^\(\)]*)+)\2\))?/,
      },
      leftMatch: {},
      attrMap: { class: "className", for: "htmlFor" },
      attrHandle: {
        href: function (e) {
          return e.getAttribute("href");
        },
        type: function (e) {
          return e.getAttribute("type");
        },
      },
      relative: {
        "+": function (C, x) {
          var z = typeof x === "string",
            B = z && !w.test(x),
            D = z && !B;
          if (B) {
            x = x.toLowerCase();
          }
          for (var y = 0, e = C.length, A; y < e; y++) {
            if ((A = C[y])) {
              while ((A = A.previousSibling) && A.nodeType !== 1) {}
              C[y] = D || (A && A.nodeName.toLowerCase() === x) ? A || false : A === x;
            }
          }
          if (D) {
            d.filter(x, C, true);
          }
        },
        ">": function (C, x) {
          var B,
            A = typeof x === "string",
            y = 0,
            e = C.length;
          if (A && !w.test(x)) {
            x = x.toLowerCase();
            for (; y < e; y++) {
              B = C[y];
              if (B) {
                var z = B.parentNode;
                C[y] = z.nodeName.toLowerCase() === x ? z : false;
              }
            }
          } else {
            for (; y < e; y++) {
              B = C[y];
              if (B) {
                C[y] = A ? B.parentNode : B.parentNode === x;
              }
            }
            if (A) {
              d.filter(x, C, true);
            }
          }
        },
        "": function (z, x, B) {
          var A,
            y = o++,
            e = t;
          if (typeof x === "string" && !w.test(x)) {
            x = x.toLowerCase();
            A = x;
            e = a;
          }
          e("parentNode", x, y, z, A, B);
        },
        "~": function (z, x, B) {
          var A,
            y = o++,
            e = t;
          if (typeof x === "string" && !w.test(x)) {
            x = x.toLowerCase();
            A = x;
            e = a;
          }
          e("previousSibling", x, y, z, A, B);
        },
      },
      find: {
        ID: function (x, y, z) {
          if (typeof y.getElementById !== "undefined" && !z) {
            var e = y.getElementById(x[1]);
            return e && e.parentNode ? [e] : [];
          }
        },
        NAME: function (y, B) {
          if (typeof B.getElementsByName !== "undefined") {
            var x = [],
              A = B.getElementsByName(y[1]);
            for (var z = 0, e = A.length; z < e; z++) {
              if (A[z].getAttribute("name") === y[1]) {
                x.push(A[z]);
              }
            }
            return x.length === 0 ? null : x;
          }
        },
        TAG: function (e, x) {
          if (typeof x.getElementsByTagName !== "undefined") {
            return x.getElementsByTagName(e[1]);
          }
        },
      },
      preFilter: {
        CLASS: function (z, x, y, e, C, D) {
          z = " " + z[1].replace(q, "") + " ";
          if (D) {
            return z;
          }
          for (var A = 0, B; (B = x[A]) != null; A++) {
            if (B) {
              if (C ^ (B.className && (" " + B.className + " ").replace(/[\t\n\r]/g, " ").indexOf(z) >= 0)) {
                if (!y) {
                  e.push(B);
                }
              } else {
                if (y) {
                  x[A] = false;
                }
              }
            }
          }
          return false;
        },
        ID: function (e) {
          return e[1].replace(q, "");
        },
        TAG: function (x, e) {
          return x[1].replace(q, "").toLowerCase();
        },
        CHILD: function (e) {
          if (e[1] === "nth") {
            if (!e[2]) {
              d.error(e[0]);
            }
            e[2] = e[2].replace(/^\+|\s*/g, "");
            var x = /(-?)(\d*)(?:n([+\-]?\d*))?/.exec(
              (e[2] === "even" && "2n") || (e[2] === "odd" && "2n+1") || (!/\D/.test(e[2]) && "0n+" + e[2]) || e[2],
            );
            e[2] = x[1] + (x[2] || 1) - 0;
            e[3] = x[3] - 0;
          } else {
            if (e[2]) {
              d.error(e[0]);
            }
          }
          e[0] = o++;
          return e;
        },
        ATTR: function (A, x, y, e, B, C) {
          var z = (A[1] = A[1].replace(q, ""));
          if (!C && k.attrMap[z]) {
            A[1] = k.attrMap[z];
          }
          A[4] = (A[4] || A[5] || "").replace(q, "");
          if (A[2] === "~=") {
            A[4] = " " + A[4] + " ";
          }
          return A;
        },
        PSEUDO: function (A, x, y, e, B) {
          if (A[1] === "not") {
            if ((n.exec(A[3]) || "").length > 1 || /^\w/.test(A[3])) {
              A[3] = d(A[3], null, null, x);
            } else {
              var z = d.filter(A[3], x, y, true ^ B);
              if (!y) {
                e.push.apply(e, z);
              }
              return false;
            }
          } else {
            if (k.match.POS.test(A[0]) || k.match.CHILD.test(A[0])) {
              return true;
            }
          }
          return A;
        },
        POS: function (e) {
          e.unshift(true);
          return e;
        },
      },
      filters: {
        enabled: function (e) {
          return e.disabled === false && e.type !== "hidden";
        },
        disabled: function (e) {
          return e.disabled === true;
        },
        checked: function (e) {
          return e.checked === true;
        },
        selected: function (e) {
          if (e.parentNode) {
            e.parentNode.selectedIndex;
          }
          return e.selected === true;
        },
        parent: function (e) {
          return !!e.firstChild;
        },
        empty: function (e) {
          return !e.firstChild;
        },
        has: function (y, x, e) {
          return !!d(e[3], y).length;
        },
        header: function (e) {
          return /h\d/i.test(e.nodeName);
        },
        text: function (y) {
          var e = y.getAttribute("type"),
            x = y.type;
          return y.nodeName.toLowerCase() === "input" && "text" === x && (e === x || e === null);
        },
        radio: function (e) {
          return e.nodeName.toLowerCase() === "input" && "radio" === e.type;
        },
        checkbox: function (e) {
          return e.nodeName.toLowerCase() === "input" && "checkbox" === e.type;
        },
        file: function (e) {
          return e.nodeName.toLowerCase() === "input" && "file" === e.type;
        },
        password: function (e) {
          return e.nodeName.toLowerCase() === "input" && "password" === e.type;
        },
        submit: function (x) {
          var e = x.nodeName.toLowerCase();
          return (e === "input" || e === "button") && "submit" === x.type;
        },
        image: function (e) {
          return e.nodeName.toLowerCase() === "input" && "image" === e.type;
        },
        reset: function (x) {
          var e = x.nodeName.toLowerCase();
          return (e === "input" || e === "button") && "reset" === x.type;
        },
        button: function (x) {
          var e = x.nodeName.toLowerCase();
          return (e === "input" && "button" === x.type) || e === "button";
        },
        input: function (e) {
          return /input|select|textarea|button/i.test(e.nodeName);
        },
        focus: function (e) {
          return e === e.ownerDocument.activeElement;
        },
      },
      setFilters: {
        first: function (x, e) {
          return e === 0;
        },
        last: function (y, x, e, z) {
          return x === z.length - 1;
        },
        even: function (x, e) {
          return e % 2 === 0;
        },
        odd: function (x, e) {
          return e % 2 === 1;
        },
        lt: function (y, x, e) {
          return x < e[3] - 0;
        },
        gt: function (y, x, e) {
          return x > e[3] - 0;
        },
        nth: function (y, x, e) {
          return e[3] - 0 === x;
        },
        eq: function (y, x, e) {
          return e[3] - 0 === x;
        },
      },
      filter: {
        PSEUDO: function (y, D, C, E) {
          var e = D[1],
            x = k.filters[e];
          if (x) {
            return x(y, C, D, E);
          } else {
            if (e === "contains") {
              return (y.textContent || y.innerText || b([y]) || "").indexOf(D[3]) >= 0;
            } else {
              if (e === "not") {
                var z = D[3];
                for (var B = 0, A = z.length; B < A; B++) {
                  if (z[B] === y) {
                    return false;
                  }
                }
                return true;
              } else {
                d.error(e);
              }
            }
          }
        },
        CHILD: function (y, A) {
          var z,
            G,
            C,
            F,
            e,
            B,
            E,
            D = A[1],
            x = y;
          switch (D) {
            case "only":
            case "first":
              while ((x = x.previousSibling)) {
                if (x.nodeType === 1) {
                  return false;
                }
              }
              if (D === "first") {
                return true;
              }
              x = y;
            case "last":
              while ((x = x.nextSibling)) {
                if (x.nodeType === 1) {
                  return false;
                }
              }
              return true;
            case "nth":
              z = A[2];
              G = A[3];
              if (z === 1 && G === 0) {
                return true;
              }
              C = A[0];
              F = y.parentNode;
              if (F && (F[i] !== C || !y.nodeIndex)) {
                B = 0;
                for (x = F.firstChild; x; x = x.nextSibling) {
                  if (x.nodeType === 1) {
                    x.nodeIndex = ++B;
                  }
                }
                F[i] = C;
              }
              E = y.nodeIndex - G;
              if (z === 0) {
                return E === 0;
              } else {
                return E % z === 0 && E / z >= 0;
              }
          }
        },
        ID: function (x, e) {
          return x.nodeType === 1 && x.getAttribute("id") === e;
        },
        TAG: function (x, e) {
          return (e === "*" && x.nodeType === 1) || (!!x.nodeName && x.nodeName.toLowerCase() === e);
        },
        CLASS: function (x, e) {
          return (" " + (x.className || x.getAttribute("class")) + " ").indexOf(e) > -1;
        },
        ATTR: function (B, z) {
          var y = z[1],
            e = d.attr ? d.attr(B, y) : k.attrHandle[y] ? k.attrHandle[y](B) : B[y] != null ? B[y] : B.getAttribute(y),
            C = e + "",
            A = z[2],
            x = z[4];
          return e == null
            ? A === "!="
            : !A && d.attr
              ? e != null
              : A === "="
                ? C === x
                : A === "*="
                  ? C.indexOf(x) >= 0
                  : A === "~="
                    ? (" " + C + " ").indexOf(x) >= 0
                    : !x
                      ? C && e !== false
                      : A === "!="
                        ? C !== x
                        : A === "^="
                          ? C.indexOf(x) === 0
                          : A === "$="
                            ? C.substr(C.length - x.length) === x
                            : A === "|="
                              ? C === x || C.substr(0, x.length + 1) === x + "-"
                              : false;
        },
        POS: function (A, x, y, B) {
          var e = x[2],
            z = k.setFilters[e];
          if (z) {
            return z(A, y, x, B);
          }
        },
      },
    });
    var j = k.match.POS,
      c = function (x, e) {
        return "\\" + (e - 0 + 1);
      };
    for (var f in k.match) {
      k.match[f] = new RegExp(k.match[f].source + /(?![^\[]*\])(?![^\(]*\))/.source);
      k.leftMatch[f] = new RegExp(/(^(?:.|\r|\n)*?)/.source + k.match[f].source.replace(/\\(\d+)/g, c));
    }
    var l = function (x, e) {
      x = Array.prototype.slice.call(x, 0);
      if (e) {
        e.push.apply(e, x);
        return e;
      }
      return x;
    };
    try {
      Array.prototype.slice.call(document.documentElement.childNodes, 0)[0].nodeType;
    } catch (v) {
      l = function (A, z) {
        var y = 0,
          x = z || [];
        if (r.call(A) === "[object Array]") {
          Array.prototype.push.apply(x, A);
        } else {
          if (typeof A.length === "number") {
            for (var e = A.length; y < e; y++) {
              x.push(A[y]);
            }
          } else {
            for (; A[y]; y++) {
              x.push(A[y]);
            }
          }
        }
        return x;
      };
    }
    var p, m;
    if (document.documentElement.compareDocumentPosition) {
      p = function (x, e) {
        if (x === e) {
          h = true;
          return 0;
        }
        if (!x.compareDocumentPosition || !e.compareDocumentPosition) {
          return x.compareDocumentPosition ? -1 : 1;
        }
        return x.compareDocumentPosition(e) & 4 ? -1 : 1;
      };
    } else {
      p = function (E, D) {
        if (E === D) {
          h = true;
          return 0;
        } else {
          if (E.sourceIndex && D.sourceIndex) {
            return E.sourceIndex - D.sourceIndex;
          }
        }
        var B,
          x,
          y = [],
          e = [],
          A = E.parentNode,
          C = D.parentNode,
          F = A;
        if (A === C) {
          return m(E, D);
        } else {
          if (!A) {
            return -1;
          } else {
            if (!C) {
              return 1;
            }
          }
        }
        while (F) {
          y.unshift(F);
          F = F.parentNode;
        }
        F = C;
        while (F) {
          e.unshift(F);
          F = F.parentNode;
        }
        B = y.length;
        x = e.length;
        for (var z = 0; z < B && z < x; z++) {
          if (y[z] !== e[z]) {
            return m(y[z], e[z]);
          }
        }
        return z === B ? m(E, e[z], -1) : m(y[z], D, 1);
      };
      m = function (x, e, y) {
        if (x === e) {
          return y;
        }
        var z = x.nextSibling;
        while (z) {
          if (z === e) {
            return -1;
          }
          z = z.nextSibling;
        }
        return 1;
      };
    }
    (function () {
      var x = document.createElement("div"),
        y = "script" + new Date().getTime(),
        e = document.documentElement;
      x.innerHTML = "<a name='" + y + "'/>";
      e.insertBefore(x, e.firstChild);
      if (document.getElementById(y)) {
        k.find.ID = function (A, B, C) {
          if (typeof B.getElementById !== "undefined" && !C) {
            var z = B.getElementById(A[1]);
            return z
              ? z.id === A[1] ||
                (typeof z.getAttributeNode !== "undefined" && z.getAttributeNode("id").nodeValue === A[1])
                ? [z]
                : undefined
              : [];
          }
        };
        k.filter.ID = function (B, z) {
          var A = typeof B.getAttributeNode !== "undefined" && B.getAttributeNode("id");
          return B.nodeType === 1 && A && A.nodeValue === z;
        };
      }
      e.removeChild(x);
      e = x = null;
    })();
    (function () {
      var e = document.createElement("div");
      e.appendChild(document.createComment(""));
      if (e.getElementsByTagName("*").length > 0) {
        k.find.TAG = function (x, B) {
          var A = B.getElementsByTagName(x[1]);
          if (x[1] === "*") {
            var z = [];
            for (var y = 0; A[y]; y++) {
              if (A[y].nodeType === 1) {
                z.push(A[y]);
              }
            }
            A = z;
          }
          return A;
        };
      }
      e.innerHTML = "<a href='#'></a>";
      if (
        e.firstChild &&
        typeof e.firstChild.getAttribute !== "undefined" &&
        e.firstChild.getAttribute("href") !== "#"
      ) {
        k.attrHandle.href = function (x) {
          return x.getAttribute("href", 2);
        };
      }
      e = null;
    })();
    if (document.querySelectorAll) {
      (function () {
        var e = d,
          z = document.createElement("div"),
          y = "__sizzle__";
        z.innerHTML = "<p class='TEST'></p>";
        if (z.querySelectorAll && z.querySelectorAll(".TEST").length === 0) {
          return;
        }
        d = function (K, B, F, J) {
          B = B || document;
          if (!J && !d.isXML(B)) {
            var I = /^(\w+$)|^\.([\w\-]+$)|^#([\w\-]+$)/.exec(K);
            if (I && (B.nodeType === 1 || B.nodeType === 9)) {
              if (I[1]) {
                return l(B.getElementsByTagName(K), F);
              } else {
                if (I[2] && k.find.CLASS && B.getElementsByClassName) {
                  return l(B.getElementsByClassName(I[2]), F);
                }
              }
            }
            if (B.nodeType === 9) {
              if (K === "body" && B.body) {
                return l([B.body], F);
              } else {
                if (I && I[3]) {
                  var E = B.getElementById(I[3]);
                  if (E && E.parentNode) {
                    if (E.id === I[3]) {
                      return l([E], F);
                    }
                  } else {
                    return l([], F);
                  }
                }
              }
              try {
                return l(B.querySelectorAll(K), F);
              } catch (G) {}
            } else {
              if (B.nodeType === 1 && B.nodeName.toLowerCase() !== "object") {
                var C = B,
                  D = B.getAttribute("id"),
                  A = D || y,
                  M = B.parentNode,
                  L = /^\s*[+~]/.test(K);
                if (!D) {
                  B.setAttribute("id", A);
                } else {
                  A = A.replace(/'/g, "\\$&");
                }
                if (L && M) {
                  B = B.parentNode;
                }
                try {
                  if (!L || M) {
                    return l(B.querySelectorAll("[id='" + A + "'] " + K), F);
                  }
                } catch (H) {
                } finally {
                  if (!D) {
                    C.removeAttribute("id");
                  }
                }
              }
            }
          }
          return e(K, B, F, J);
        };
        for (var x in e) {
          d[x] = e[x];
        }
        z = null;
      })();
    }
    (function () {
      var e = document.documentElement,
        y = e.matchesSelector || e.mozMatchesSelector || e.webkitMatchesSelector || e.msMatchesSelector;
      if (y) {
        var A = !y.call(document.createElement("div"), "div"),
          x = false;
        try {
          y.call(document.documentElement, "[test!='']:sizzle");
        } catch (z) {
          x = true;
        }
        d.matchesSelector = function (C, E) {
          E = E.replace(/\=\s*([^'"\]]*)\s*\]/g, "='$1']");
          if (!d.isXML(C)) {
            try {
              if (x || (!k.match.PSEUDO.test(E) && !/!=/.test(E))) {
                var B = y.call(C, E);
                if (B || !A || (C.document && C.document.nodeType !== 11)) {
                  return B;
                }
              }
            } catch (D) {}
          }
          return d(E, null, null, [C]).length > 0;
        };
      }
    })();
    (function () {
      var e = document.createElement("div");
      e.innerHTML = "<div class='test e'></div><div class='test'></div>";
      if (!e.getElementsByClassName || e.getElementsByClassName("e").length === 0) {
        return;
      }
      e.lastChild.className = "e";
      if (e.getElementsByClassName("e").length === 1) {
        return;
      }
      k.order.splice(1, 0, "CLASS");
      k.find.CLASS = function (x, y, z) {
        if (typeof y.getElementsByClassName !== "undefined" && !z) {
          return y.getElementsByClassName(x[1]);
        }
      };
      e = null;
    })();
    function a(x, C, B, F, D, E) {
      for (var z = 0, y = F.length; z < y; z++) {
        var e = F[z];
        if (e) {
          var A = false;
          e = e[x];
          while (e) {
            if (e[i] === B) {
              A = F[e.sizset];
              break;
            }
            if (e.nodeType === 1 && !E) {
              e[i] = B;
              e.sizset = z;
            }
            if (e.nodeName.toLowerCase() === C) {
              A = e;
              break;
            }
            e = e[x];
          }
          F[z] = A;
        }
      }
    }
    function t(x, C, B, F, D, E) {
      for (var z = 0, y = F.length; z < y; z++) {
        var e = F[z];
        if (e) {
          var A = false;
          e = e[x];
          while (e) {
            if (e[i] === B) {
              A = F[e.sizset];
              break;
            }
            if (e.nodeType === 1) {
              if (!E) {
                e[i] = B;
                e.sizset = z;
              }
              if (typeof C !== "string") {
                if (e === C) {
                  A = true;
                  break;
                }
              } else {
                if (d.filter(C, [e]).length > 0) {
                  A = e;
                  break;
                }
              }
            }
            e = e[x];
          }
          F[z] = A;
        }
      }
    }
    if (document.documentElement.contains) {
      d.contains = function (x, e) {
        return x !== e && (x.contains ? x.contains(e) : true);
      };
    } else {
      if (document.documentElement.compareDocumentPosition) {
        d.contains = function (x, e) {
          return !!(x.compareDocumentPosition(e) & 16);
        };
      } else {
        d.contains = function () {
          return false;
        };
      }
    }
    d.isXML = function (e) {
      var x = (e ? e.ownerDocument || e : 0).documentElement;
      return x ? x.nodeName !== "HTML" : false;
    };
    var s = function (y, e, C) {
      var B,
        D = [],
        A = "",
        E = e.nodeType ? [e] : e;
      while ((B = k.match.PSEUDO.exec(y))) {
        A += B[0];
        y = y.replace(k.match.PSEUDO, "");
      }
      y = k.relative[y] ? y + "*" : y;
      for (var z = 0, x = E.length; z < x; z++) {
        d(y, E[z], D, C);
      }
      return d.filter(A, D);
    };
    baidu.dom.query = d;
  })();
  baidu.dom.extend({
    removeClass: function (b) {
      if (arguments.length <= 0) {
        baidu.forEach(this, function (c) {
          c.className = "";
        });
      }
      switch (typeof b) {
        case "string":
          b = String(b).replace(/^\s+/g, "").replace(/\s+$/g, "").replace(/\s+/g, " ");
          var a = b.split(" ");
          baidu.forEach(this, function (d) {
            var e = d.className;
            for (var c = 0; c < a.length; c++) {
              while ((" " + e + " ").indexOf(" " + a[c] + " ") >= 0) {
                e = (" " + e + " ").replace(" " + a[c] + " ", " ");
              }
            }
            d.className = e.replace(/^\s+/g, "").replace(/\s+$/g, "");
          });
          break;
        case "function":
          baidu.forEach(this, function (e, c, d) {
            baidu.dom(e).removeClass(b.call(e, c, e.className));
          });
          break;
      }
      return this;
    },
  });
  baidu.dom.extend({
    removeData: (function () {
      var a = baidu.key,
        b = baidu.global("_maps_HTMLElementData");
      return function (c) {
        baidu.forEach(this, function (d) {
          !d[a] && (d[a] = baidu.id());
        });
        baidu.forEach(this, function (e) {
          var d = b[e[a]];
          if (typeof c == "string") {
            d && delete d[c];
          } else {
            if (baidu.type(c) == "array") {
              baidu.forEach(c, function (f) {
                d && delete d[f];
              });
            }
          }
        });
        return this;
      };
    })(),
  });
  baidu.dom.extend({
    removeProp: function (c) {
      if (arguments.length <= 0 || !c || typeof c !== "string") {
        return this;
      }
      var b = baidu.dom,
        a = baidu._util_;
      c = a.propFix[c] || c;
      baidu.forEach(this, function (d) {
        try {
          d[c] = undefined;
          delete d[c];
        } catch (f) {}
      });
      return this;
    },
  });
  baidu.dom.removeStyle = (function () {
    var b = document.createElement("DIV"),
      a,
      c = baidu.dom._g;
    if (b.style.removeProperty) {
      a = function (e, d) {
        e = c(e);
        e.style.removeProperty(d);
        return e;
      };
    } else {
      if (b.style.removeAttribute) {
        a = function (e, d) {
          e = c(e);
          e.style.removeAttribute(baidu.string.toCamelCase(d));
          return e;
        };
      }
    }
    b = null;
    return a;
  })();
  baidu.object.each = function (e, c) {
    var b, a, d;
    if ("function" == typeof c) {
      for (a in e) {
        if (e.hasOwnProperty(a)) {
          d = e[a];
          b = c.call(e, d, a);
          if (b === false) {
            break;
          }
        }
      }
    }
    return e;
  };
  baidu.dom.extend({
    setStyles: function (b) {
      element = this[0];
      for (var a in b) {
        baidu.dom.setStyle(element, a, b[a]);
      }
      return element;
    },
  });
  baidu.dom._styleFilter[baidu.dom._styleFilter.length] = {
    set: function (a, b) {
      if (b.constructor == Number && !/zIndex|fontWeight|opacity|zoom|lineHeight/i.test(a)) {
        b = b + "px";
      }
      return b;
    },
  };
  baidu.event.getTarget = function (a) {
    a.originalEvent && (a = a.originalEvent);
    return a.target || a.srcElement;
  };
  baidu.dom.setBorderBoxSize = function (c, b) {
    var a = {};
    b.width && (a.width = parseFloat(b.width));
    b.height && (a.height = parseFloat(b.height));
    function d(f, e) {
      return parseFloat(baidu.dom.getStyle(f, e)) || 0;
    }
    if (baidu.browser.isStrict) {
      if (b.width) {
        a.width =
          parseFloat(b.width) -
          d(c, "paddingLeft") -
          d(c, "paddingRight") -
          d(c, "borderLeftWidth") -
          d(c, "borderRightWidth");
        a.width < 0 && (a.width = 0);
      }
      if (b.height) {
        a.height =
          parseFloat(b.height) -
          d(c, "paddingTop") -
          d(c, "paddingBottom") -
          d(c, "borderTopWidth") -
          d(c, "borderBottomWidth");
        a.height < 0 && (a.height = 0);
      }
    }
    return baidu.dom.setStyles(c, a);
  };
  baidu.dom.setOuterHeight = baidu.dom.setBorderBoxHeight = function (b, a) {
    return baidu.dom.setBorderBoxSize(b, { height: a });
  };
  baidu.dom.setOuterWidth = baidu.dom.setBorderBoxWidth = function (a, b) {
    return baidu.dom.setBorderBoxSize(a, { width: b });
  };
  baidu.dom.resizable = function (d, g) {
    var z,
      m,
      i = {},
      c,
      a = {},
      r,
      x,
      u,
      b,
      e,
      k,
      o,
      s = false,
      j = false,
      v = {
        direction: ["e", "s", "se"],
        minWidth: 16,
        minHeight: 16,
        classPrefix: "tangram",
        directionHandlePosition: {},
      };
    if (!(z = baidu.dom.g(d)) && baidu.dom.getStyle(z, "position") == "static") {
      return false;
    }
    b = z.offsetParent;
    var n = baidu.dom.getStyle(z, "position");
    m = baidu.extend(v, g);
    baidu.forEach(["minHeight", "minWidth", "maxHeight", "maxWidth"], function (A) {
      m[A] && (m[A] = parseFloat(m[A]));
    });
    r = [m.minWidth || 0, m.maxWidth || Number.MAX_VALUE, m.minHeight || 0, m.maxHeight || Number.MAX_VALUE];
    y();
    function y() {
      k = baidu.extend(
        {
          e: { right: "-5px", top: "0px", width: "7px", height: z.offsetHeight },
          s: { left: "0px", bottom: "-5px", height: "7px", width: z.offsetWidth },
          n: { left: "0px", top: "-5px", height: "7px", width: z.offsetWidth },
          w: { left: "-5px", top: "0px", height: z.offsetHeight, width: "7px" },
          se: { right: "1px", bottom: "1px", height: "16px", width: "16px" },
          sw: { left: "1px", bottom: "1px", height: "16px", width: "16px" },
          ne: { right: "1px", top: "1px", height: "16px", width: "16px" },
          nw: { left: "1px", top: "1px", height: "16px", width: "16px" },
        },
        m.directionHandlePosition,
      );
      baidu.forEach(m.direction, function (A) {
        var B = m.classPrefix.split(" ");
        B[0] = B[0] + "-resizable-" + A;
        var D = baidu.dom.create("div", { className: B.join(" ") }),
          C = k[A];
        C["cursor"] = A + "-resize";
        C["position"] = "absolute";
        baidu.dom.setStyles(D, C);
        D.key = A;
        D.style.MozUserSelect = "none";
        z.appendChild(D);
        i[A] = D;
        baidu.dom(D).on("mousedown", h);
      });
      s = false;
    }
    function f() {
      e && t();
      baidu.object.each(i, function (A) {
        baidu.dom(A).off("mousedown", h);
        baidu.dom.remove(A);
      });
      s = true;
    }
    function l(A) {
      if (!s) {
        m = baidu.extend(m, A || {});
        f();
        y();
      }
    }
    function h(C) {
      j && t();
      var B = baidu.event.getTarget(C),
        A = B.key;
      e = B;
      j = true;
      if (B.setCapture) {
        B.setCapture();
      } else {
        if (window.captureEvents) {
          window.captureEvents(Event.MOUSEMOVE | Event.MOUSEUP);
        }
      }
      u = baidu.dom.getStyle(document.body, "cursor");
      baidu.dom.setStyle(document.body, "cursor", A + "-resize");
      var E = baidu.dom(document.body);
      E.on("mouseup", t);
      E.on("selectstart", p);
      x = document.body.style.MozUserSelect;
      document.body.style.MozUserSelect = "none";
      var D = baidu.page.getMousePosition();
      a = q();
      o = setInterval(function () {
        w(A, D);
      }, 20);
      baidu.isFunction(m.onresizestart) && m.onresizestart();
      baidu.event.preventDefault(C);
    }
    function t() {
      if (e && e.releaseCapture) {
        e.releaseCapture();
      } else {
        if (window.releaseEvents) {
          window.releaseEvents(Event.MOUSEMOVE | Event.MOUSEUP);
        }
      }
      baidu.dom(document.body).off("mouseup", t);
      baidu.dom(document).off("selectstart", p);
      document.body.style.MozUserSelect = x;
      baidu.dom(document.body).off("selectstart", p);
      clearInterval(o);
      baidu.dom.setStyle(document.body, "cursor", u);
      e = null;
      j = false;
      baidu.isFunction(m.onresizeend) && m.onresizeend();
    }
    function w(B, H) {
      var G = baidu.page.getMousePosition(),
        C = a["width"],
        A = a["height"],
        F = a["top"],
        E = a["left"],
        D;
      if (B.indexOf("e") >= 0) {
        C = Math.max(G.x - H.x + a["width"], r[0]);
        C = Math.min(C, r[1]);
      } else {
        if (B.indexOf("w") >= 0) {
          C = Math.max(H.x - G.x + a["width"], r[0]);
          C = Math.min(C, r[1]);
          E -= C - a["width"];
        }
      }
      if (B.indexOf("s") >= 0) {
        A = Math.max(G.y - H.y + a["height"], r[2]);
        A = Math.min(A, r[3]);
      } else {
        if (B.indexOf("n") >= 0) {
          A = Math.max(H.y - G.y + a["height"], r[2]);
          A = Math.min(A, r[3]);
          F -= A - a["height"];
        }
      }
      D = { width: C, height: A, top: F, left: E };
      baidu.dom.setOuterHeight(z, A);
      baidu.dom.setOuterWidth(z, C);
      baidu.dom.setStyles(z, { top: F, left: E });
      i["n"] && baidu.dom.setStyle(i["n"], "width", C);
      i["s"] && baidu.dom.setStyle(i["s"], "width", C);
      i["e"] && baidu.dom.setStyle(i["e"], "height", A);
      i["w"] && baidu.dom.setStyle(i["w"], "height", A);
      baidu.isFunction(m.onresize) && m.onresize({ current: D, original: a });
    }
    function p(A) {
      return baidu.event.preventDefault(A, false);
    }
    function q() {
      var A = baidu.dom.getPosition(z.offsetParent),
        B = baidu.dom.getPosition(z),
        D,
        C;
      if (n == "absolute") {
        D = B.top - (z.offsetParent == document.body ? 0 : A.top);
        C = B.left - (z.offsetParent == document.body ? 0 : A.left);
      } else {
        D = parseFloat(baidu.dom.getStyle(z, "top")) || -parseFloat(baidu.dom.getStyle(z, "bottom")) || 0;
        C = parseFloat(baidu.dom.getStyle(z, "left")) || -parseFloat(baidu.dom.getStyle(z, "right")) || 0;
      }
      baidu.dom.setStyles(z, { top: D, left: C });
      return { width: z.offsetWidth, height: z.offsetHeight, top: D, left: C };
    }
    return { cancel: f, update: l, enable: y };
  };
  baidu._util_.smartScroll = function (d) {
    var a = { scrollLeft: "pageXOffset", scrollTop: "pageYOffset" }[d],
      e = d === "scrollLeft",
      b = {};
    function f(g) {
      return g && g.nodeType === 9;
    }
    function c(g) {
      return baidu.type(g) == "Window" ? g : f(g) ? g.defaultView || g.parentWindow : false;
    }
    return {
      get: function (g) {
        var h = c(g);
        return h
          ? a in h
            ? h[a]
            : (baidu.browser.isStrict && h.document.documentElement[d]) || h.document.body[d]
          : g[d];
      },
      set: function (g, i) {
        if (!g) {
          return;
        }
        var h = c(g);
        h ? h.scrollTo(e ? i : this.get(g), !e ? i : this.get(g)) : (g[d] = i);
      },
    };
  };
  baidu.dom.extend({
    scrollLeft: (function () {
      var a = baidu._util_.smartScroll("scrollLeft");
      return function (b) {
        b && baidu.check("^(?:number|string)$", "baidu.dom.scrollLeft");
        if (this.size() <= 0) {
          return b === undefined ? 0 : this;
        }
        return b === undefined ? a.get(this[0]) : a.set(this[0], b) || this;
      };
    })(),
  });
  baidu.dom.extend({
    scrollTop: (function () {
      var a = baidu._util_.smartScroll("scrollTop");
      return function (b) {
        b && baidu.check("^(?:number|string)$", "baidu.dom.scrollTop");
        if (this.size() <= 0) {
          return b === undefined ? 0 : this;
        }
        return b === undefined ? a.get(this[0]) : a.set(this[0], b) || this;
      };
    })(),
  });
  baidu.dom.extend({
    setAttrs: function (a) {
      element = this[0];
      for (var b in a) {
        baidu.dom.setAttr(element, b, a[b]);
      }
      return element;
    },
  });
  baidu.dom.setPixel = function (b, a, c) {
    typeof c != "undefined" && (baidu.dom.g(b).style[a] = c + (!isNaN(c) ? "px" : ""));
  };
  baidu.dom.setPosition = function (d, a) {
    var c = {
      left: a.left - (parseFloat(baidu.dom.getStyle(d, "margin-left")) || 0),
      top: a.top - (parseFloat(baidu.dom.getStyle(d, "margin-top")) || 0),
    };
    d = baidu.dom.g(d);
    for (var b in c) {
      baidu.dom.setStyle(d, b, c[b]);
    }
    return d;
  };
  baidu.dom.extend({
    show: function () {
      baidu._util_.showHide(this, true);
      return this;
    },
  });
  baidu.dom.extend({
    siblings: function (a) {
      var b = [];
      baidu.forEach(this, function (e) {
        var d = [],
          f = [],
          c = e;
        while ((c = c.previousSibling)) {
          c.nodeType == 1 && d.push(c);
        }
        while ((e = e.nextSibling)) {
          e.nodeType == 1 && f.push(e);
        }
        baidu.merge(b, d.reverse().concat(f));
      });
      return baidu.dom(baidu.dom.match(b, a));
    },
  });
  baidu.dom.extend({
    slice: (function () {
      var a = Array.prototype.slice;
      return function (c, b) {
        baidu.check("number(,number)?", "baidu.dom.slice");
        return baidu.dom(a.apply(this, arguments));
      };
    })(),
  });
  baidu.dom.toggle = function (a) {
    a = baidu.dom.g(a);
    a.style.display = a.style.display == "none" ? "" : "none";
    return a;
  };
  baidu.dom.extend({
    toggleClass: function (d, b) {
      var c = typeof d;
      var b = typeof b === "undefined" ? b : Boolean(b);
      if (arguments.length <= 0) {
        baidu.forEach(this, function (e) {
          e.className = "";
        });
      }
      switch (typeof d) {
        case "string":
          d = d.replace(/^\s+/g, "").replace(/\s+$/g, "").replace(/\s+/g, " ");
          var a = d.split(" ");
          baidu.forEach(this, function (f) {
            var g = f.className;
            for (var e = 0; e < a.length; e++) {
              if ((" " + g + " ").indexOf(" " + a[e] + " ") > -1 && typeof b === "undefined") {
                g = (" " + g + " ").replace(" " + a[e] + " ", " ");
              } else {
                if ((" " + g + " ").indexOf(" " + a[e] + " ") === -1 && typeof b === "undefined") {
                  g += " " + a[e];
                } else {
                  if ((" " + g + " ").indexOf(" " + a[e] + " ") === -1 && b === true) {
                    g += " " + a[e];
                  } else {
                    if ((" " + g + " ").indexOf(" " + a[e] + " ") > -1 && b === false) {
                      g = g.replace(a[e], "");
                    }
                  }
                }
              }
            }
            f.className = g.replace(/^\s+/g, "").replace(/\s+$/g, "");
          });
          break;
        case "function":
          baidu.forEach(this, function (f, e) {
            baidu.dom(f).toggleClass(d.call(f, e, f.className), b);
          });
          break;
      }
      return this;
    },
  });
  baidu.dom.toggleClass = function (a, b) {
    if (baidu.dom.hasClass(a, b)) {
      baidu.dom.removeClass(a, b);
    } else {
      baidu.dom.addClass(a, b);
    }
  };
  baidu.dom.extend({
    trigger: (function () {
      var i = baidu._util_.eventBase;
      var a = /msie/i.test(navigator.userAgent);
      var p = { keydown: 1, keyup: 1, keypress: 1 };
      var b = {
        click: 1,
        dblclick: 1,
        mousedown: 1,
        mousemove: 1,
        mouseup: 1,
        mouseover: 1,
        mouseout: 1,
        mouseenter: 1,
        mouseleave: 1,
        contextmenu: 1,
      };
      var l = {
        abort: 1,
        blur: 1,
        change: 1,
        error: 1,
        focus: 1,
        focusin: 1,
        focusout: 1,
        load: 1,
        unload: 1,
        reset: 1,
        resize: 1,
        scroll: 1,
        select: 1,
        submit: 1,
      };
      var j = { scroll: 1, resize: 1, reset: 1, submit: 1, change: 1, select: 1, error: 1, abort: 1 };
      var f = { submit: 1 };
      var o = {
        KeyEvents: ["bubbles", "cancelable", "view", "ctrlKey", "altKey", "shiftKey", "metaKey", "keyCode", "charCode"],
        MouseEvents: [
          "bubbles",
          "cancelable",
          "view",
          "detail",
          "screenX",
          "screenY",
          "clientX",
          "clientY",
          "ctrlKey",
          "altKey",
          "shiftKey",
          "metaKey",
          "button",
          "relatedTarget",
        ],
        HTMLEvents: ["bubbles", "cancelable"],
        UIEvents: ["bubbles", "cancelable", "view", "detail"],
        Events: ["bubbles", "cancelable"],
      };
      baidu.extend(j, p);
      baidu.extend(j, b);
      var h = function (r) {
        return r.replace(/^\w/, function (t) {
          return t.toUpperCase();
        });
      };
      var m = function (u) {
        var r = [],
          t = 0,
          s;
        for (s in u) {
          if (u.hasOwnProperty(s)) {
            r[t++] = u[s];
          }
        }
        return r;
      };
      var e = function (v, t) {
        var s = 0,
          r = v.length,
          u = {};
        for (; s < r; s++) {
          u[v[s]] = t[v[s]];
          delete t[v[s]];
        }
        return u;
      };
      var g = function (t, s, r) {
        r = baidu.extend({}, r);
        var u = m(e(o[s], r)),
          v = document.createEvent(s);
        u.unshift(t);
        switch (s) {
          case "KeyEvents":
            v.initKeyEvent.apply(v, u);
            break;
          case "MouseEvents":
            v.initMouseEvent.apply(v, u);
            break;
          case "UIEvents":
            v.initUIEvent.apply(v, u);
            break;
          default:
            v.initEvent.apply(v, u);
            break;
        }
        if (r.triggerData) {
          v.triggerData = r.triggerData;
        }
        baidu.extend(v, r);
        return v;
      };
      var d = function (r) {
        var s;
        if (document.createEventObject) {
          s = document.createEventObject();
          baidu.extend(s, r);
        }
        return s;
      };
      var k = function (s, r) {
        r = e(o["KeyEvents"], r);
        var u;
        if (document.createEvent) {
          try {
            u = g(s, "KeyEvents", r);
          } catch (t) {
            try {
              u = g(s, "Events", r);
            } catch (t) {
              u = g(s, "UIEvents", r);
            }
          }
        } else {
          r.keyCode = r.charCode > 0 ? r.charCode : r.keyCode;
          u = d(r);
        }
        return u;
      };
      var q = function (s, r) {
        r = e(o["MouseEvents"], r);
        var t;
        if (document.createEvent) {
          t = g(s, "MouseEvents", r);
          if (r.relatedTarget && !t.relatedTarget) {
            if ("mouseout" == s.toLowerCase()) {
              t.toElement = r.relatedTarget;
            } else {
              if ("mouseover" == s.toLowerCase()) {
                t.fromElement = r.relatedTarget;
              }
            }
          }
        } else {
          r.button = r.button == 0 ? 1 : r.button == 1 ? 4 : baidu.lang.isNumber(r.button) ? r.button : 0;
          t = d(r);
        }
        return t;
      };
      var n = function (s, r) {
        r.bubbles = j.hasOwnProperty(s);
        r = e(o["HTMLEvents"], r);
        var u;
        if (document.createEvent) {
          try {
            u = g(s, "HTMLEvents", r);
          } catch (t) {
            try {
              u = g(s, "UIEvents", r);
            } catch (t) {
              u = g(s, "Events", r);
            }
          }
        } else {
          u = d(r);
        }
        return u;
      };
      var c = function (s, u, t) {
        var w;
        var w = {
          bubbles: true,
          cancelable: true,
          view: window,
          detail: 1,
          screenX: 0,
          screenY: 0,
          clientX: 0,
          clientY: 0,
          ctrlKey: false,
          altKey: false,
          shiftKey: false,
          metaKey: false,
          keyCode: 0,
          charCode: 0,
          button: 0,
          relatedTarget: null,
        };
        if (p[u]) {
          w = k(u, w);
        } else {
          if (b[u]) {
            w = q(u, w);
          } else {
            if (l[u]) {
              w = n(u, w);
            } else {
              return baidu(s).triggerHandler(u, t);
            }
          }
        }
        if (w) {
          if (t) {
            w.triggerData = t;
          }
          var r;
          if (s.dispatchEvent) {
            r = s.dispatchEvent(w);
          } else {
            if (s.fireEvent) {
              r = s.fireEvent("on" + u, w);
            }
          }
          if (r !== false && f[u]) {
            try {
              if (s[u]) {
                s[u]();
              } else {
                if (((u = h(u)), s[u])) {
                  s[u]();
                }
              }
            } catch (v) {}
          }
        }
      };
      return function (s, r) {
        this.each(function () {
          c(this, s, r);
        });
        return this;
      };
    })(),
  });
  baidu.dom.extend({
    unbind: function (b, a) {
      return this.off(b, a);
    },
  });
  baidu.dom.extend({
    undelegate: function (a, c, b) {
      return this.off(c, a, b);
    },
  });
  baidu.dom.extend({
    unique: function (a) {
      return baidu.dom(baidu.array(this.toArray()).unique(a));
    },
  });
  baidu.dom._matchNode = function (a, c, d) {
    a = baidu.dom.g(a);
    for (var b = a[d]; b; b = b[c]) {
      if (b.nodeType == 1) {
        return b;
      }
    }
    return null;
  };
  baidu.dom._styleFilter[baidu.dom._styleFilter.length] = {
    get: function (c, d) {
      if (/color/i.test(c) && d.indexOf("rgb(") != -1) {
        var e = d.split(",");
        d = "#";
        for (var b = 0, a; (a = e[b]); b++) {
          a = parseInt(a.replace(/[^\d]/gi, ""), 10).toString(16);
          d += a.length == 1 ? "0" + a : a;
        }
        d = d.toUpperCase();
      }
      return d;
    },
  };
  baidu.dom._styleFixer.display =
    baidu.browser.ie && baidu.browser.ie < 8
      ? {
          set: function (a, b) {
            a = a.style;
            if (b == "inline-block") {
              a.display = "inline";
              a.zoom = 1;
            } else {
              a.display = b;
            }
          },
        }
      : baidu.browser.firefox && baidu.browser.firefox < 3
        ? {
            set: function (a, b) {
              a.style.display = b == "inline-block" ? "-moz-inline-box" : b;
            },
          }
        : null;
  baidu.dom._styleFixer["float"] = baidu.browser.ie ? "styleFloat" : "cssFloat";
  baidu.dom._styleFixer.opacity = baidu.browser.ie
    ? {
        get: function (a) {
          var b = a.style.filter;
          return b && b.indexOf("opacity=") >= 0 ? parseFloat(b.match(/opacity=([^)]*)/)[1]) / 100 + "" : "1";
        },
        set: function (a, c) {
          var b = a.style;
          b.filter =
            (b.filter || "").replace(/alpha\([^\)]*\)/gi, "") + (c == 1 ? "" : "alpha(opacity=" + c * 100 + ")");
          b.zoom = 1;
        },
      }
    : null;
  baidu.dom._styleFixer.width = baidu.dom._styleFixer.height = {
    get: function (b, a, c) {
      var a = a.replace(/^[a-z]/, function (e) {
          return e.toUpperCase();
        }),
        d = b["client" + a] || b["offset" + a];
      return d > 0 ? d + "px" : !c || c == "auto" ? 0 + "px" : d;
    },
    set: function (b, c, a) {
      b.style[a] = c;
    },
  };
  baidu.dom._styleFixer.textOverflow = (function () {
    var b = {};
    function a(e) {
      var f = e.length;
      if (f > 0) {
        f = e[f - 1];
        e.length--;
      } else {
        f = null;
      }
      return f;
    }
    function c(e, f) {
      e[baidu.browser.firefox ? "textContent" : "innerText"] = f;
    }
    function d(m, h, s) {
      var k = baidu.browser.ie ? m.currentStyle || m.style : getComputedStyle(m, null),
        r = k.fontWeight,
        q =
          "font-family:" +
          k.fontFamily +
          ";font-size:" +
          k.fontSize +
          ";word-spacing:" +
          k.wordSpacing +
          ";font-weight:" +
          ((parseInt(r) || 0) == 401 ? 700 : r) +
          ";font-style:" +
          k.fontStyle +
          ";font-variant:" +
          k.fontVariant,
        e = b[q];
      if (!e) {
        k = m.appendChild(document.createElement("div"));
        k.style.cssText = "float:left;" + q;
        e = b[q] = [];
        for (var n = 0; n < 256; n++) {
          n == 32 ? (k.innerHTML = "&nbsp;") : c(k, String.fromCharCode(n));
          e[n] = k.offsetWidth;
        }
        c(k, "\u4e00");
        e[256] = k.offsetWidth;
        c(k, "\u4e00\u4e00");
        e[257] = k.offsetWidth - e[256] * 2;
        e[258] = e[".".charCodeAt(0)] * 3 + e[257] * 3;
        m.removeChild(k);
      }
      for (var l = m.firstChild, p = e[256], g = e[257], f = e[258], u = [], s = s ? f : 0; l; l = l.nextSibling) {
        if (h < s) {
          m.removeChild(l);
        } else {
          if (l.nodeType == 3) {
            for (var n = 0, t = l.nodeValue, j = t.length; n < j; n++) {
              k = t.charCodeAt(n);
              u[u.length] = [h, l, n];
              h -= (n ? g : 0) + (k < 256 ? e[k] : p);
              if (h < s) {
                break;
              }
            }
          } else {
            k = l.tagName;
            if (k == "IMG" || k == "TABLE") {
              k = l;
              l = l.previousSibling;
              m.removeChild(k);
            } else {
              u[u.length] = [h, l];
              h -= l.offsetWidth;
            }
          }
        }
      }
      if (h < s) {
        while ((k = a(u))) {
          h = k[0];
          l = k[1];
          k = k[2];
          if (l.nodeType == 3) {
            if (h >= f) {
              l.nodeValue = l.nodeValue.substring(0, k) + "...";
              return true;
            } else {
              if (!k) {
                m.removeChild(l);
              }
            }
          } else {
            if (d(l, h, true)) {
              return true;
            } else {
              m.removeChild(l);
            }
          }
        }
        m.innerHTML = "";
      }
    }
    return {
      get: function (g) {
        var f = baidu.browser,
          e = dom.getStyle;
        return (f.opera ? e("OTextOverflow") : f.firefox ? g._baiduOverflow : e("textOverflow")) || "clip";
      },
      set: function (f, h) {
        var e = baidu.browser;
        if (f.tagName == "TD" || f.tagName == "TH" || e.firefox) {
          f._baiduHTML && (f.innerHTML = f._baiduHTML);
          if (h == "ellipsis") {
            f._baiduHTML = f.innerHTML;
            var i = document.createElement("div"),
              g = f.appendChild(i).offsetWidth;
            f.removeChild(i);
            d(f, g);
          } else {
            f._baiduHTML = "";
          }
        }
        i = f.style;
        e.opera ? (i.OTextOverflow = h) : e.firefox ? (f._baiduOverflow = h) : (i.textOverflow = h);
      },
    };
  })();
  baidu.lang.isArray = baidu.isArray;
  baidu.lang.toArray = function (b) {
    if (b === null || b === undefined) {
      return [];
    }
    if (baidu.lang.isArray(b)) {
      return b;
    }
    if (typeof b.length !== "number" || typeof b === "string" || baidu.lang.isFunction(b)) {
      return [b];
    }
    if (b.item) {
      var a = b.length,
        c = new Array(a);
      while (a--) {
        c[a] = b[a];
      }
      return c;
    }
    return [].slice.call(b);
  };
  baidu.fn.extend({
    methodize: function (a) {
      var b = this.fn;
      return function () {
        return b.apply(this, [a ? this[a] : this].concat([].slice.call(arguments)));
      };
    },
  });
  baidu.fn.extend({
    wrapReturnValue: function (c, b) {
      var a = this.fn;
      b = b | 0;
      return function () {
        var d = a.apply(this, arguments);
        if (!b) {
          return new c(d);
        }
        if (b > 0) {
          return new c(arguments[b - 1]);
        }
        return d;
      };
    },
  });
  baidu.fn.wrapReturnValue = function (a, c, b) {
    return baidu.fn(a).wrapReturnValue(c, b);
  };
  baidu.fn.extend({
    multize: function (b, a) {
      var d = this.fn;
      function c() {
        var l = arguments[0],
          h = b ? c : d,
          f = [],
          k = Array.prototype.slice.call(arguments, 0),
          e;
        if (l instanceof Array) {
          for (var g = 0, j; (j = l[g]); g++) {
            k[0] = j;
            e = h.apply(this, k);
            if (a) {
              e && (f = f.concat(e));
            } else {
              f.push(e);
            }
          }
          return f;
        } else {
          return d.apply(this, arguments);
        }
      }
      return c;
    },
  });
  baidu.fn.multize = function (c, b, a) {
    return baidu.fn(c).multize(b, a);
  };
  baidu.element = function (b) {
    var a = baidu.dom._g(b);
    if (!a && baidu.dom.query) {
      a = baidu.dom.query(b);
    }
    return new baidu.element.Element(a);
  };
  baidu.element.Element = function (a) {
    if (!baidu.element._init) {
      baidu.element._makeChain();
      baidu.element._init = true;
    }
    this._dom = (a.tagName || "").toLowerCase() == "select" ? [a] : baidu.lang.toArray(a);
  };
  baidu.element._toChainFunction = function (c, b, a) {
    return baidu.fn.methodize(baidu.fn.wrapReturnValue(baidu.fn.multize(c, 0, 1), baidu.element.Element, b), "_dom");
  };
  baidu.element._makeChain = function () {
    var b = baidu.element.Element.prototype,
      c = baidu.element._toChainFunction;
    baidu.forEach("draggable droppable resizable fixable".split(" "), function (d) {
      b[d] = c(baidu.dom[d], 1);
    });
    baidu.forEach(
      "remove getText contains getAttr getPosition getStyle hasClass intersect hasAttr getComputedStyle".split(" "),
      function (d) {
        b[d] = b[d.replace(/^get[A-Z]/g, a)] = c(baidu.dom[d], -1);
      },
    );
    baidu.forEach(
      (
        "addClass empty hide show insertAfter insertBefore insertHTML removeClass " +
        "setAttr setAttrs setStyle setStyles show toggleClass toggle next first " +
        "getAncestorByClass getAncestorBy getAncestorByTag getDocument getParent getWindow " +
        "last next prev g removeStyle setBorderBoxSize setOuterWidth setOuterHeight " +
        "setBorderBoxWidth setBorderBoxHeight setPosition children query"
      ).split(" "),
      function (d) {
        b[d] = b[d.replace(/^get[A-Z]/g, a)] = c(baidu.dom[d], 0);
      },
    );
    b["q"] = b["Q"] = c(function (e, d) {
      return baidu.dom.q.apply(this, [d, e].concat([].slice.call(arguments, 2)));
    }, 0);
    baidu.forEach("on un".split(" "), function (d) {
      b[d] = c(baidu.event[d], 0);
    });
    baidu.forEach(
      (
        "blur focus focusin focusout load resize scroll unload click dblclick " +
        "mousedown mouseup mousemove mouseover mouseout mouseenter mouseleave " +
        "change select submit keydown keypress keyup error"
      ).split(" "),
      function (d) {
        b[d] = function (e) {
          return this.on(d, e);
        };
      },
    );
    function a(d) {
      return d.charAt(3).toLowerCase();
    }
  };
  baidu.element.extend = function (a) {
    var b = baidu.element;
    baidu.object.each(a, function (d, c) {
      b.Element.prototype[c] = baidu.element._toChainFunction(d, -1);
    });
  };
  baidu.event.EventArg = function (c, e) {
    e = e || window;
    c = c || e.event;
    var d = e.document;
    this.target = c.target || c.srcElement;
    this.keyCode = c.which || c.keyCode;
    for (var a in c) {
      var b = c[a];
      if ("function" != typeof b) {
        this[a] = b;
      }
    }
    if (!this.pageX && this.pageX !== 0) {
      this.pageX = (c.clientX || 0) + (d.documentElement.scrollLeft || d.body.scrollLeft);
      this.pageY = (c.clientY || 0) + (d.documentElement.scrollTop || d.body.scrollTop);
    }
    this._event = c;
  };
  baidu.event.EventArg.prototype.preventDefault = function () {
    if (this._event.preventDefault) {
      this._event.preventDefault();
    } else {
      this._event.returnValue = false;
    }
    return this;
  };
  baidu.event.EventArg.prototype.stopPropagation = function () {
    if (this._event.stopPropagation) {
      this._event.stopPropagation();
    } else {
      this._event.cancelBubble = true;
    }
    return this;
  };
  baidu.event.EventArg.prototype.stop = function () {
    return this.stopPropagation().preventDefault();
  };
  baidu.object.values = function (d) {
    var a = [],
      c = 0,
      b;
    for (b in d) {
      if (d.hasOwnProperty(b)) {
        a[c++] = d[b];
      }
    }
    return a;
  };
  baidu.lang.isNumber = baidu.isNumber;
  baidu.event.fire = (function () {
    var d = baidu.browser,
      k = { keydown: 1, keyup: 1, keypress: 1 },
      a = { click: 1, dblclick: 1, mousedown: 1, mousemove: 1, mouseup: 1, mouseover: 1, mouseout: 1 },
      h = {
        abort: 1,
        blur: 1,
        change: 1,
        error: 1,
        focus: 1,
        load: d.ie ? 0 : 1,
        reset: 1,
        resize: 1,
        scroll: 1,
        select: 1,
        submit: 1,
        unload: d.ie ? 0 : 1,
      },
      f = { scroll: 1, resize: 1, reset: 1, submit: 1, change: 1, select: 1, error: 1, abort: 1 },
      j = {
        KeyEvents: ["bubbles", "cancelable", "view", "ctrlKey", "altKey", "shiftKey", "metaKey", "keyCode", "charCode"],
        MouseEvents: [
          "bubbles",
          "cancelable",
          "view",
          "detail",
          "screenX",
          "screenY",
          "clientX",
          "clientY",
          "ctrlKey",
          "altKey",
          "shiftKey",
          "metaKey",
          "button",
          "relatedTarget",
        ],
        HTMLEvents: ["bubbles", "cancelable"],
        UIEvents: ["bubbles", "cancelable", "view", "detail"],
        Events: ["bubbles", "cancelable"],
      };
    baidu.object.extend(f, k);
    baidu.object.extend(f, a);
    function c(q, o) {
      var n = 0,
        m = q.length,
        p = {};
      for (; n < m; n++) {
        p[q[n]] = o[q[n]];
        delete o[q[n]];
      }
      return p;
    }
    function e(o, n, m) {
      m = baidu.object.extend({}, m);
      var p = baidu.object.values(c(j[n], m)),
        q = document.createEvent(n);
      p.unshift(o);
      if ("KeyEvents" == n) {
        q.initKeyEvent.apply(q, p);
      } else {
        if ("MouseEvents" == n) {
          q.initMouseEvent.apply(q, p);
        } else {
          if ("UIEvents" == n) {
            q.initUIEvent.apply(q, p);
          } else {
            q.initEvent.apply(q, p);
          }
        }
      }
      baidu.object.extend(q, m);
      return q;
    }
    function b(m) {
      var n;
      if (document.createEventObject) {
        n = document.createEventObject();
        baidu.object.extend(n, m);
      }
      return n;
    }
    function g(p, m) {
      m = c(j["KeyEvents"], m);
      var q;
      if (document.createEvent) {
        try {
          q = e(p, "KeyEvents", m);
        } catch (o) {
          try {
            q = e(p, "Events", m);
          } catch (n) {
            q = e(p, "UIEvents", m);
          }
        }
      } else {
        m.keyCode = m.charCode > 0 ? m.charCode : m.keyCode;
        q = b(m);
      }
      return q;
    }
    function l(n, m) {
      m = c(j["MouseEvents"], m);
      var o;
      if (document.createEvent) {
        o = e(n, "MouseEvents", m);
        if (m.relatedTarget && !o.relatedTarget) {
          if ("mouseout" == n.toLowerCase()) {
            o.toElement = m.relatedTarget;
          } else {
            if ("mouseover" == n.toLowerCase()) {
              o.fromElement = m.relatedTarget;
            }
          }
        }
      } else {
        m.button = m.button == 0 ? 1 : m.button == 1 ? 4 : baidu.lang.isNumber(m.button) ? m.button : 0;
        o = b(m);
      }
      return o;
    }
    function i(o, m) {
      m.bubbles = f.hasOwnProperty(o);
      m = c(j["HTMLEvents"], m);
      var q;
      if (document.createEvent) {
        try {
          q = e(o, "HTMLEvents", m);
        } catch (p) {
          try {
            q = e(o, "UIEvents", m);
          } catch (n) {
            q = e(o, "Events", m);
          }
        }
      } else {
        q = b(m);
      }
      return q;
    }
    return function (n, o, m) {
      var p;
      o = o.replace(/^on/i, "");
      n = baidu.dom._g(n);
      m = baidu.object.extend(
        {
          bubbles: true,
          cancelable: true,
          view: window,
          detail: 1,
          screenX: 0,
          screenY: 0,
          clientX: 0,
          clientY: 0,
          ctrlKey: false,
          altKey: false,
          shiftKey: false,
          metaKey: false,
          keyCode: 0,
          charCode: 0,
          button: 0,
          relatedTarget: null,
        },
        m,
      );
      if (k[o]) {
        p = g(o, m);
      } else {
        if (a[o]) {
          p = l(o, m);
        } else {
          if (h[o]) {
            p = i(o, m);
          } else {
            throw new Error(o + " is not support!");
          }
        }
      }
      if (p) {
        if (n.dispatchEvent) {
          n.dispatchEvent(p);
        } else {
          if (n.fireEvent) {
            n.fireEvent("on" + o, p);
          }
        }
      }
    };
  })();
  baidu.event.get = function (a, b) {
    return new baidu.event.EventArg(a, b);
  };
  baidu.event.getEvent = function (b) {
    if (window.event) {
      return window.event;
    } else {
      var c = arguments.callee,
        a;
      do {
        a = c.arguments[0];
        if (a && (/Event/.test(a) || a.originalEvent)) {
          return a.originalEvent || a;
        }
      } while ((c = c.caller));
      return null;
    }
  };
  baidu.event.getKeyCode = function (a) {
    a.originalEvent && (a = a.originalEvent);
    return a.which || a.keyCode;
  };
  baidu.event.getPageX = function (b) {
    b.originalEvent && (b = b.originalEvent);
    var a = b.pageX,
      c = document;
    if (!a && a !== 0) {
      a = (b.clientX || 0) + (c.documentElement.scrollLeft || c.body.scrollLeft);
    }
    return a;
  };
  baidu.event.getPageY = function (b) {
    b.originalEvent && (b = b.originalEvent);
    var a = b.pageY,
      c = document;
    if (!a && a !== 0) {
      a = (b.clientY || 0) + (c.documentElement.scrollTop || c.body.scrollTop);
    }
    return a;
  };
  baidu.event.once = function (a, b, c) {
    return baidu.dom(baidu.dom._g(a)).one(b, c)[0];
  };
  baidu.event.stopPropagation = function (a) {
    a.originalEvent && (a = a.originalEvent);
    if (a.stopPropagation) {
      a.stopPropagation();
    } else {
      a.cancelBubble = true;
    }
  };
  baidu.event.stop = function (a) {
    a.originalEvent && (a = a.originalEvent);
    baidu.event.stopPropagation(a);
    baidu.event.preventDefault(a);
  };
  baidu.fn.abstractMethod = function () {
    throw Error("unimplemented abstract method");
  };
  baidu.fn.extend({
    bind: function (a) {
      var b = this.fn,
        c = arguments.length > 1 ? Array.prototype.slice.call(arguments, 1) : null;
      return function () {
        var e = baidu.type(b) === "string" ? a[b] : b,
          d = c ? c.concat(Array.prototype.slice.call(arguments, 0)) : arguments;
        return e.apply(a || e, d);
      };
    },
  });
  baidu.fn.bind = function (c, b) {
    var a = baidu.fn(c);
    return a.bind.apply(a, Array.prototype.slice.call(arguments, 1));
  };
  baidu.lang.register = function (b, d, a) {
    var c = b["\x06r"] || (b["\x06r"] = []);
    c[c.length] = d;
    for (var e in a) {
      b.prototype[e] = a[e];
    }
  };
  baidu.createChain(
    "form",
    function (a) {
      return typeof a === "undefined" ? new baidu.form.$Form() : new baidu.form.$Form(a);
    },
    function (a) {
      this.form = a;
    },
  );
  baidu.form.extend({
    json: function (e) {
      var c = this.form;
      var b = c.elements,
        e =
          e ||
          function (q, i) {
            return q;
          },
        h = {},
        o,
        k,
        p,
        d,
        a,
        n,
        m,
        l;
      function f(i, q) {
        var r = h[i];
        if (r) {
          r.push || (h[i] = [r]);
          h[i].push(q);
        } else {
          h[i] = q;
        }
      }
      for (var g = 0, j = b.length; g < j; g++) {
        o = b[g];
        p = o.name;
        if (!o.disabled && p) {
          k = o.type;
          d = baidu.url.escapeSymbol(o.value);
          switch (k) {
            case "radio":
            case "checkbox":
              if (!o.checked) {
                break;
              }
            case "textarea":
            case "text":
            case "password":
            case "hidden":
            case "file":
            case "select-one":
              f(p, e(d, p));
              break;
            case "select-multiple":
              a = o.options;
              m = a.length;
              for (n = 0; n < m; n++) {
                l = a[n];
                if (l.selected) {
                  f(p, e(l.value, p));
                }
              }
              break;
          }
        }
      }
      return h;
    },
  });
  baidu.form.extend({
    serialize: function (e) {
      var c = this.form;
      var b = c.elements,
        e =
          e ||
          function (q, i) {
            return q;
          },
        h = [],
        o,
        k,
        p,
        d,
        a,
        n,
        m,
        l;
      function f(i, q) {
        h.push(i + "=" + q);
      }
      for (var g = 0, j = b.length; g < j; g++) {
        o = b[g];
        p = o.name;
        if (!o.disabled && p) {
          k = o.type;
          d = baidu.url.escapeSymbol(o.value);
          switch (k) {
            case "radio":
            case "checkbox":
              if (!o.checked) {
                break;
              }
            case "textarea":
            case "text":
            case "password":
            case "hidden":
            case "file":
            case "select-one":
              f(p, e(d, p));
              break;
            case "select-multiple":
              a = o.options;
              m = a.length;
              for (n = 0; n < m; n++) {
                l = a[n];
                if (l.selected) {
                  f(p, e(l.value, p));
                }
              }
              break;
          }
        }
      }
      return h;
    },
  });
  baidu.fx = baidu.fx || {};
  baidu.lang.inherits = function (g, e, d) {
    var c,
      f,
      a = g.prototype,
      b = new Function();
    b.prototype = e.prototype;
    f = g.prototype = new b();
    for (c in a) {
      f[c] = a[c];
    }
    g.prototype.constructor = g;
    g.superClass = e.prototype;
    typeof d == "string" && (f.__type = d);
    g.extend = function (j) {
      for (var h in j) {
        f[h] = j[h];
      }
      return g;
    };
    return g;
  };
  baidu.fx.Timeline = function (a) {
    baidu.lang.Class.call(this);
    this.interval = 16;
    this.duration = 500;
    this.dynamic = true;
    baidu.object.extend(this, a);
  };
  baidu.lang.inherits(baidu.fx.Timeline, baidu.lang.Class, "baidu.fx.Timeline").extend({
    launch: function () {
      var a = this;
      a.dispatchEvent("onbeforestart");
      typeof a.initialize == "function" && a.initialize();
      a["\x06btime"] = new Date().getTime();
      a["\x06etime"] = a["\x06btime"] + (a.dynamic ? a.duration : 0);
      a["\x06pulsed"]();
      return a;
    },
    "\x06pulsed": function () {
      var b = this;
      var a = new Date().getTime();
      b.percent = (a - b["\x06btime"]) / b.duration;
      b.dispatchEvent("onbeforeupdate");
      if (a >= b["\x06etime"]) {
        typeof b.render == "function" && b.render(b.transition((b.percent = 1)));
        typeof b.finish == "function" && b.finish();
        b.dispatchEvent("onafterfinish");
        b.dispose();
        return;
      }
      typeof b.render == "function" && b.render(b.transition(b.percent));
      b.dispatchEvent("onafterupdate");
      b["\x06timer"] = setTimeout(function () {
        b["\x06pulsed"]();
      }, b.interval);
    },
    transition: function (a) {
      return a;
    },
    cancel: function () {
      this["\x06timer"] && clearTimeout(this["\x06timer"]);
      this["\x06etime"] = this["\x06btime"];
      typeof this.restore == "function" && this.restore();
      this.dispatchEvent("oncancel");
      this.dispose();
    },
    end: function () {
      this["\x06timer"] && clearTimeout(this["\x06timer"]);
      this["\x06etime"] = this["\x06btime"];
      this["\x06pulsed"]();
    },
  });
  baidu.fx.create = function (d, b, c) {
    var e = new baidu.fx.Timeline(b);
    e.element = d;
    e.__type = c || e.__type;
    e["\x06original"] = {};
    var a = "baidu_current_effect";
    e.addEventListener("onbeforestart", function () {
      var g = this,
        f;
      g.attribName = "att_" + g.__type.replace(/\W/g, "_");
      f = g.element.getAttribute(a);
      g.element.setAttribute(a, (f || "") + "|" + g.guid + "|", 0);
      if (!g.overlapping) {
        (f = g.element.getAttribute(g.attribName)) && window[baidu.guid]._instances[f].cancel();
        g.element.setAttribute(g.attribName, g.guid, 0);
      }
    });
    e["\x06clean"] = function (h) {
      var g = this,
        f;
      if ((h = g.element)) {
        h.removeAttribute(g.attribName);
        f = h.getAttribute(a);
        f = f.replace("|" + g.guid + "|", "");
        if (!f) {
          h.removeAttribute(a);
        } else {
          h.setAttribute(a, f, 0);
        }
      }
    };
    e.addEventListener("oncancel", function () {
      this["\x06clean"]();
      this["\x06restore"]();
    });
    e.addEventListener("onafterfinish", function () {
      this["\x06clean"]();
      this.restoreAfterFinish && this["\x06restore"]();
    });
    e.protect = function (f) {
      this["\x06original"][f] = this.element.style[f];
    };
    e["\x06restore"] = function () {
      var j = this["\x06original"],
        h = this.element.style,
        f;
      for (var g in j) {
        f = j[g];
        if (typeof f == "undefined") {
          continue;
        }
        h[g] = f;
        if (!f && h.removeAttribute) {
          h.removeAttribute(g);
        } else {
          if (!f && h.removeProperty) {
            h.removeProperty(g);
          }
        }
      }
    };
    return e;
  };
  baidu.fx.collapse = function (c, b) {
    if (!(c = baidu.dom.g(c))) {
      return null;
    }
    var h = c,
      f,
      a,
      g = {
        vertical: {
          value: "height",
          offset: "offsetHeight",
          stylesValue: ["paddingBottom", "paddingTop", "borderTopWidth", "borderBottomWidth"],
        },
        horizontal: {
          value: "width",
          offset: "offsetWidth",
          stylesValue: ["paddingLeft", "paddingRight", "borderLeftWidth", "borderRightWidth"],
        },
      };
    var d = baidu.fx.create(
      h,
      baidu.object.extend(
        {
          orientation: "vertical",
          initialize: function () {
            a = g[this.orientation];
            this.protect(a.value);
            this.protect("overflow");
            this.restoreAfterFinish = true;
            f = h[a.offset];
            h.style.overflow = "hidden";
          },
          transition: function (e) {
            return Math.pow(1 - e, 2);
          },
          render: function (e) {
            h.style[a.value] = Math.floor(e * f) + "px";
          },
          finish: function () {
            baidu.dom.hide(h);
          },
        },
        b || {},
      ),
      "baidu.fx.expand_collapse",
    );
    return d.launch();
  };
  baidu.lang.instance = (function () {
    var a = baidu.global("_maps_id");
    return function (b) {
      return a[b] || null;
    };
  })();
  baidu.fx.current = function (d) {
    if (!(d = baidu.dom.g(d))) {
      return null;
    }
    var b,
      f,
      e = /\|([^\|]+)\|/g;
    do {
      if ((f = d.getAttribute("baidu_current_effect"))) {
        break;
      }
    } while ((d = d.parentNode) && d.nodeType == 1);
    if (!f) {
      return null;
    }
    if ((b = f.match(e))) {
      e = /\|([^\|]+)\|/;
      for (var c = 0; c < b.length; c++) {
        e.test(b[c]);
        b[c] = baidu.lang.instance(RegExp["\x241"]);
      }
    }
    return b;
  };
  baidu.fx.expand = function (c, b) {
    if (!(c = baidu.dom.g(c))) {
      return null;
    }
    var h = c,
      f,
      a,
      g = {
        vertical: {
          value: "height",
          offset: "offsetHeight",
          stylesValue: ["paddingBottom", "paddingTop", "borderTopWidth", "borderBottomWidth"],
        },
        horizontal: {
          value: "width",
          offset: "offsetWidth",
          stylesValue: ["paddingLeft", "paddingRight", "borderLeftWidth", "borderRightWidth"],
        },
      };
    var d = baidu.fx.create(
      h,
      baidu.object.extend(
        {
          orientation: "vertical",
          initialize: function () {
            a = g[this.orientation];
            baidu.dom.show(h);
            this.protect(a.value);
            this.protect("overflow");
            this.restoreAfterFinish = true;
            f = h[a.offset];
            function e(k, j) {
              var i = parseInt(baidu.dom.getStyle(k, j));
              i = isNaN(i) ? 0 : i;
              i = baidu.lang.isNumber(i) ? i : 0;
              return i;
            }
            baidu.forEach(a.stylesValue, function (i) {
              f -= e(h, i);
            });
            h.style.overflow = "hidden";
            h.style[a.value] = "1px";
          },
          transition: function (e) {
            return Math.sqrt(e);
          },
          render: function (e) {
            h.style[a.value] = Math.floor(e * f) + "px";
          },
        },
        b || {},
      ),
      "baidu.fx.expand_collapse",
    );
    return d.launch();
  };
  baidu.fx.opacity = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    a = baidu.object.extend({ from: 0, to: 1 }, a || {});
    var d = b;
    var c = baidu.fx.create(
      d,
      baidu.object.extend(
        {
          initialize: function () {
            baidu.dom.show(b);
            if (baidu.browser.ie) {
              this.protect("filter");
            } else {
              this.protect("opacity");
              this.protect("KHTMLOpacity");
            }
            this.distance = this.to - this.from;
          },
          render: function (e) {
            var f = this.distance * e + this.from;
            if (!baidu.browser.ie) {
              d.style.opacity = f;
              d.style.KHTMLOpacity = f;
            } else {
              d.style.filter = "progid:DXImageTransform.Microsoft.Alpha(opacity:" + Math.floor(f * 100) + ")";
            }
          },
        },
        a,
      ),
      "baidu.fx.opacity",
    );
    return c.launch();
  };
  baidu.fx.fadeIn = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    var c = baidu.fx.opacity(b, baidu.object.extend({ from: 0, to: 1, restoreAfterFinish: true }, a || {}));
    c.__type = "baidu.fx.fadeIn";
    return c;
  };
  baidu.fx.fadeOut = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    var c = baidu.fx.opacity(b, baidu.object.extend({ from: 1, to: 0, restoreAfterFinish: true }, a || {}));
    c.addEventListener("onafterfinish", function () {
      baidu.dom.hide(this.element);
    });
    c.__type = "baidu.fx.fadeOut";
    return c;
  };
  baidu.fx.getTransition = function (c) {
    var b = baidu.fx.transitions;
    if (!c || typeof b[c] != "string") {
      c = "linear";
    }
    return new Function("percent", b[c]);
  };
  baidu.fx.transitions = {
    none: "return 0",
    full: "return 1",
    linear: "return percent",
    reverse: "return 1 - percent",
    parabola: "return Math.pow(percent, 2)",
    antiparabola: "return 1 - Math.pow(1 - percent, 2)",
    sinoidal: "return (-Math.cos(percent * Math.PI)/2) + 0.5",
    wobble: "return (-Math.cos(percent * Math.PI * (9 * percent))/2) + 0.5",
    spring: "return 1 - (Math.cos(percent * 4.5 * Math.PI) * Math.exp(-percent * 6))",
  };
  baidu.string.extend({
    formatColor: (function () {
      var c = /^\#[\da-f]{6}$/i,
        b = /^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/i,
        a = {
          black: "#000000",
          silver: "#c0c0c0",
          gray: "#808080",
          white: "#ffffff",
          maroon: "#800000",
          red: "#ff0000",
          purple: "#800080",
          fuchsia: "#ff00ff",
          green: "#008000",
          lime: "#00ff00",
          olive: "#808000",
          yellow: "#ffff0",
          navy: "#000080",
          blue: "#0000ff",
          teal: "#008080",
          aqua: "#00ffff",
        };
      return function () {
        var e = this.valueOf();
        if (c.test(e)) {
          return e;
        } else {
          if (b.test(e)) {
            for (var j, h = 1, e = "#"; h < 4; h++) {
              j = parseInt(RegExp["\x24" + h]).toString(16);
              e += ("00" + j).substr(j.length);
            }
            return e;
          } else {
            if (/^\#[\da-f]{3}$/.test(e)) {
              var g = e.charAt(1),
                f = e.charAt(2),
                d = e.charAt(3);
              return "#" + g + g + f + f + d + d;
            } else {
              if (a[e]) {
                return a[e];
              }
            }
          }
        }
        return "";
      };
    })(),
  });
  baidu.fx.highlight = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    var d = b;
    var c = baidu.fx.create(
      d,
      baidu.object.extend(
        {
          initialize: function () {
            var k = this,
              j = baidu.dom.getStyle,
              h = baidu.string.formatColor,
              f = h(j(d, "color")) || "#000000",
              e = h(j(d, "backgroundColor"));
            k.beginColor = k.beginColor || e || "#FFFF00";
            k.endColor = k.endColor || e || "#FFFFFF";
            k.finalColor = k.finalColor || k.endColor || k.element.style.backgroundColor;
            k.textColor == f && (k.textColor = "");
            this.protect("color");
            this.protect("backgroundColor");
            k.c_b = [];
            k.c_d = [];
            k.t_b = [];
            k.t_d = [];
            for (var l, g = 0; g < 3; g++) {
              l = 2 * g + 1;
              k.c_b[g] = parseInt(k.beginColor.substr(l, 2), 16);
              k.c_d[g] = parseInt(k.endColor.substr(l, 2), 16) - k.c_b[g];
              if (k.textColor) {
                k.t_b[g] = parseInt(f.substr(l, 2), 16);
                k.t_d[g] = parseInt(k.textColor.substr(l, 2), 16) - k.t_b[g];
              }
            }
          },
          render: function (j) {
            for (var h = this, f = "#", e = "#", k, g = 0; g < 3; g++) {
              k = Math.round(h.c_b[g] + h.c_d[g] * j).toString(16);
              f += ("00" + k).substr(k.length);
              if (h.textColor) {
                k = Math.round(h.t_b[g] + h.t_d[g] * j).toString(16);
                e += ("00" + k).substr(k.length);
              }
            }
            d.style.backgroundColor = f;
            h.textColor && (d.style.color = e);
          },
          finish: function () {
            this.textColor && (d.style.color = this.textColor);
            d.style.backgroundColor = this.finalColor;
          },
        },
        a || {},
      ),
      "baidu.fx.highlight",
    );
    return c.launch();
  };
  baidu.fx.mask = function (c, a) {
    if (!(c = baidu.dom.g(c)) || baidu.dom.getStyle(c, "position") != "absolute") {
      return null;
    }
    var g = c,
      b = {};
    a = a || {};
    var f = /^(\d+px|\d?\d(\.\d+)?%|100%|left|center|right)(\s+(\d+px|\d?\d(\.\d+)?%|100%|top|center|bottom))?/i;
    !f.test(a.startOrigin) && (a.startOrigin = "0px 0px");
    var a = baidu.object.extend({ restoreAfterFinish: true, from: 0, to: 1 }, a || {});
    var d = baidu.fx.create(
      g,
      baidu.object.extend(
        {
          initialize: function () {
            g.style.display = "";
            this.protect("clip");
            b.width = g.offsetWidth;
            b.height = g.offsetHeight;
            f.test(this.startOrigin);
            var k = RegExp["\x241"].toLowerCase(),
              j = RegExp["\x244"].toLowerCase(),
              i = this.element.offsetWidth,
              l = this.element.offsetHeight,
              h,
              e;
            if (/\d+%/.test(k)) {
              h = (parseInt(k, 10) / 100) * i;
            } else {
              if (/\d+px/.test(k)) {
                h = parseInt(k);
              } else {
                if (k == "left") {
                  h = 0;
                } else {
                  if (k == "center") {
                    h = i / 2;
                  } else {
                    if (k == "right") {
                      h = i;
                    }
                  }
                }
              }
            }
            if (!j) {
              e = l / 2;
            } else {
              if (/\d+%/.test(j)) {
                e = (parseInt(j, 10) / 100) * l;
              } else {
                if (/\d+px/.test(j)) {
                  e = parseInt(j);
                } else {
                  if (j == "top") {
                    e = 0;
                  } else {
                    if (j == "center") {
                      e = l / 2;
                    } else {
                      if (j == "bottom") {
                        e = l;
                      }
                    }
                  }
                }
              }
            }
            b.x = h;
            b.y = e;
          },
          render: function (k) {
            var l = this.to * k + this.from * (1 - k),
              j = b.y * (1 - l) + "px ",
              i = b.x * (1 - l) + "px ",
              h = b.x * (1 - l) + b.width * l + "px ",
              e = b.y * (1 - l) + b.height * l + "px ";
            g.style.clip = "rect(" + j + h + e + i + ")";
          },
          finish: function () {
            if (this.to < this.from) {
              g.style.display = "none";
            }
          },
        },
        a,
      ),
      "baidu.fx.mask",
    );
    return d.launch();
  };
  baidu.fx.move = function (b, a) {
    if (!(b = baidu.dom.g(b)) || baidu.dom.getStyle(b, "position") == "static") {
      return null;
    }
    a = baidu.object.extend({ x: 0, y: 0 }, a || {});
    if (a.x == 0 && a.y == 0) {
      return null;
    }
    var c = baidu.fx.create(
      b,
      baidu.object.extend(
        {
          initialize: function () {
            this.protect("top");
            this.protect("left");
            this.originX = parseInt(baidu.dom.getStyle(b, "left")) || 0;
            this.originY = parseInt(baidu.dom.getStyle(b, "top")) || 0;
          },
          transition: function (d) {
            return 1 - Math.pow(1 - d, 2);
          },
          render: function (d) {
            b.style.top = this.y * d + this.originY + "px";
            b.style.left = this.x * d + this.originX + "px";
          },
        },
        a,
      ),
      "baidu.fx.move",
    );
    return c.launch();
  };
  baidu.fx.moveBy = function (b, f, a) {
    if (!(b = baidu.dom.g(b)) || baidu.dom.getStyle(b, "position") == "static" || typeof f != "object") {
      return null;
    }
    var e = {};
    e.x = f[0] || f.x || 0;
    e.y = f[1] || f.y || 0;
    var c = baidu.fx.move(b, baidu.object.extend(e, a || {}));
    return c;
  };
  baidu.fx.moveTo = function (d, b, c) {
    if (!(d = baidu.dom.g(d)) || baidu.dom.getStyle(d, "position") == "static" || typeof b != "object") {
      return null;
    }
    var f = [b[0] || b.x || 0, b[1] || b.y || 0];
    var a = parseInt(baidu.dom.getStyle(d, "left")) || 0;
    var g = parseInt(baidu.dom.getStyle(d, "top")) || 0;
    var e = baidu.fx.move(d, baidu.object.extend({ x: f[0] - a, y: f[1] - g }, c || {}));
    return e;
  };
  baidu.fx.scale = function (c, a) {
    if (!(c = baidu.dom.g(c))) {
      return null;
    }
    a = baidu.object.extend({ from: 0.1, to: 1 }, a || {});
    var e = /^(-?\d+px|\d?\d(\.\d+)?%|100%|left|center|right)(\s+(-?\d+px|\d?\d(\.\d+)?%|100%|top|center|bottom))?/i;
    !e.test(a.transformOrigin) && (a.transformOrigin = "0px 0px");
    var b = {},
      d = baidu.fx.create(
        c,
        baidu.object.extend(
          {
            fade: true,
            initialize: function () {
              baidu.dom.show(c);
              var k = this,
                f = b,
                p = c.style,
                j = function (o) {
                  k.protect(o);
                };
              if (baidu.browser.ie) {
                j("top");
                j("left");
                j("position");
                j("zoom");
                j("filter");
                this.offsetX = parseInt(baidu.dom.getStyle(c, "left")) || 0;
                this.offsetY = parseInt(baidu.dom.getStyle(c, "top")) || 0;
                if (baidu.dom.getStyle(c, "position") == "static") {
                  p.position = "relative";
                }
                e.test(this.transformOrigin);
                var i = RegExp["\x241"].toLowerCase(),
                  h = RegExp["\x244"].toLowerCase(),
                  l = this.element.offsetWidth,
                  g = this.element.offsetHeight,
                  n,
                  m;
                if (/\d+%/.test(i)) {
                  n = (parseInt(i, 10) / 100) * l;
                } else {
                  if (/\d+px/.test(i)) {
                    n = parseInt(i);
                  } else {
                    if (i == "left") {
                      n = 0;
                    } else {
                      if (i == "center") {
                        n = l / 2;
                      } else {
                        if (i == "right") {
                          n = l;
                        }
                      }
                    }
                  }
                }
                if (!h) {
                  m = g / 2;
                } else {
                  if (/\d+%/.test(h)) {
                    m = (parseInt(h, 10) / 100) * g;
                  } else {
                    if (/\d+px/.test(h)) {
                      m = parseInt(h);
                    } else {
                      if (h == "top") {
                        m = 0;
                      } else {
                        if (h == "center") {
                          m = g / 2;
                        } else {
                          if (h == "bottom") {
                            m = g;
                          }
                        }
                      }
                    }
                  }
                }
                p.zoom = this.from;
                f.cx = n;
                f.cy = m;
              } else {
                j("WebkitTransform");
                j("WebkitTransformOrigin");
                j("MozTransform");
                j("MozTransformOrigin");
                j("OTransform");
                j("OTransformOrigin");
                j("transform");
                j("transformOrigin");
                j("opacity");
                j("KHTMLOpacity");
                p.WebkitTransform = p.MozTransform = p.OTransform = p.transform = "scale(" + this.from + ")";
                p.WebkitTransformOrigin =
                  p.MozTransformOrigin =
                  p.OTransformOrigin =
                  p.transformOrigin =
                    this.transformOrigin;
              }
            },
            render: function (i) {
              var g = c.style,
                f = this.to == 1,
                f = typeof this.opacityTrend == "boolean" ? this.opacityTrend : f,
                h = f ? this.percent : 1 - this.percent,
                j = this.to * i + this.from * (1 - i);
              if (baidu.browser.ie) {
                g.zoom = j;
                if (this.fade) {
                  g.filter = "progid:DXImageTransform.Microsoft.Alpha(opacity:" + Math.floor(h * 100) + ")";
                }
                g.top = this.offsetY + b.cy * (1 - j);
                g.left = this.offsetX + b.cx * (1 - j);
              } else {
                g.WebkitTransform = g.MozTransform = g.OTransform = g.transform = "scale(" + j + ")";
                if (this.fade) {
                  g.KHTMLOpacity = g.opacity = h;
                }
              }
            },
          },
          a,
        ),
        "baidu.fx.scale",
      );
    return d.launch();
  };
  baidu.fx.zoomOut = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    a = baidu.object.extend(
      {
        to: 0.1,
        from: 1,
        opacityTrend: false,
        restoreAfterFinish: true,
        transition: function (d) {
          return 1 - Math.pow(1 - d, 2);
        },
      },
      a || {},
    );
    var c = baidu.fx.scale(b, a);
    c.addEventListener("onafterfinish", function () {
      baidu.dom.hide(this.element);
    });
    return c;
  };
  baidu.fx.puff = function (b, a) {
    return baidu.fx.zoomOut(b, baidu.object.extend({ to: 1.8, duration: 800, transformOrigin: "50% 50%" }, a || {}));
  };
  baidu.fx.pulsate = function (c, a, b) {
    if (!(c = baidu.dom.g(c))) {
      return null;
    }
    if (isNaN(a) || a == 0) {
      return null;
    }
    var f = c;
    var d = baidu.fx.create(
      f,
      baidu.object.extend(
        {
          initialize: function () {
            this.protect("visibility");
          },
          transition: function (e) {
            return Math.cos(2 * Math.PI * e);
          },
          render: function (e) {
            f.style.visibility = e > 0 ? "visible" : "hidden";
          },
          finish: function () {
            setTimeout(function () {
              baidu.fx.pulsate(c, --a, b);
            }, 10);
          },
        },
        b,
      ),
      "baidu.fx.pulsate",
    );
    return d.launch();
  };
  baidu.fx.remove = function (b, a) {
    var c = a.onafterfinish ? a.onafterfinish : new Function();
    return baidu.fx.fadeOut(
      b,
      baidu.object.extend(a || {}, {
        onafterfinish: function () {
          baidu.dom.remove(this.element);
          c.call(this);
        },
      }),
    );
  };
  baidu.fx.scrollBy = function (b, g, a) {
    if (!(b = baidu.dom.g(b)) || typeof g != "object") {
      return null;
    }
    var f = {},
      e = {};
    f.x = g[0] || g.x || 0;
    f.y = g[1] || g.y || 0;
    var c = baidu.fx.create(
      b,
      baidu.object.extend(
        {
          initialize: function () {
            var h = (e.sTop = b.scrollTop);
            var d = (e.sLeft = b.scrollLeft);
            e.sx = Math.min(b.scrollWidth - b.clientWidth - d, f.x);
            e.sy = Math.min(b.scrollHeight - b.clientHeight - h, f.y);
          },
          transition: function (d) {
            return 1 - Math.pow(1 - d, 2);
          },
          render: function (d) {
            b.scrollTop = e.sy * d + e.sTop;
            b.scrollLeft = e.sx * d + e.sLeft;
          },
          restore: function () {
            b.scrollTop = e.sTop;
            b.scrollLeft = e.sLeft;
          },
        },
        a,
      ),
      "baidu.fx.scroll",
    );
    return c.launch();
  };
  baidu.fx.scrollTo = function (c, a, b) {
    if (!(c = baidu.dom.g(c)) || typeof a != "object") {
      return null;
    }
    var e = {};
    e.x = (a[0] || a.x || 0) - c.scrollLeft;
    e.y = (a[1] || a.y || 0) - c.scrollTop;
    return baidu.fx.scrollBy(c, e, b);
  };
  baidu.fx.shake = function (b, g, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    var f = b;
    g = g || [];
    function c() {
      for (var e = 0; e < arguments.length; e++) {
        if (!isNaN(arguments[e])) {
          return arguments[e];
        }
      }
    }
    var d = baidu.fx.create(
      f,
      baidu.object.extend(
        {
          initialize: function () {
            this.protect("top");
            this.protect("left");
            this.protect("position");
            this.restoreAfterFinish = true;
            if (baidu.dom.getStyle(f, "position") == "static") {
              f.style.position = "relative";
            }
            var e = this["\x06original"];
            this.originX = parseInt(e.left || 0);
            this.originY = parseInt(e.top || 0);
            this.offsetX = c(g[0], g.x, 16);
            this.offsetY = c(g[1], g.y, 5);
          },
          transition: function (h) {
            var e = 1 - h;
            return Math.floor(e * 16) % 2 == 1 ? e : h - 1;
          },
          render: function (e) {
            f.style.top = this.offsetY * e + this.originY + "px";
            f.style.left = this.offsetX * e + this.originX + "px";
          },
        },
        a || {},
      ),
      "baidu.fx.shake",
    );
    return d.launch();
  };
  baidu.fx.zoomIn = function (b, a) {
    if (!(b = baidu.dom.g(b))) {
      return null;
    }
    a = baidu.object.extend(
      {
        to: 1,
        from: 0.1,
        restoreAfterFinish: true,
        transition: function (c) {
          return Math.pow(c, 2);
        },
      },
      a || {},
    );
    return baidu.fx.scale(b, a);
  };
  baidu._util_.smartAjax =
    baidu._util_.smartAjax ||
    function (a) {
      return function (b, d, e, c) {
        if (baidu.type(d) === "function") {
          c = c || e;
          e = d;
          d = undefined;
        }
        baidu.ajax({ type: a, url: b, data: d, success: e, dataType: c });
      };
    };
  baidu.get = baidu.get || baidu._util_.smartAjax("get");
  baidu.global.get = function (a) {
    return baidu.global(a);
  };
  baidu.global.set = function (b, c, a) {
    return baidu.global(b, c, a);
  };
  baidu.global.getZIndex = function (a, c) {
    var b = baidu.global.get("zIndex");
    if (a) {
      b[a] = b[a] + (c || 1);
    }
    return b[a];
  };
  baidu.global.set("zIndex", { popup: 50000, dialog: 1000 }, true);
  baidu.i18n = baidu.i18n || {};
  baidu.i18n.cultures = baidu.i18n.cultures || {};
  baidu.i18n.cultures["en-US"] = baidu.object.extend(baidu.i18n.cultures["en-US"] || {}, {
    calendar: {
      dateFormat: "yyyy-MM-dd",
      titleNames: "#{MM}&nbsp;#{yyyy}",
      monthNames: [
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
      ],
      monthNamesShort: ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"],
      dayNames: { mon: "Mon", tue: "Tue", wed: "Wed", thu: "Thu", fri: "Fri", sat: "Sat", sun: "Sun" },
    },
    timeZone: -5,
    whitespace: new RegExp("(^[\\s\\t\\xa0\\u3000]+)|([\\u3000\\xa0\\s\\t]+\x24)", "g"),
    number: {
      group: ",",
      groupLength: 3,
      decimal: ".",
      positive: "",
      negative: "-",
      _format: function (b, a) {
        return baidu.i18n.number._format(b, {
          group: this.group,
          groupLength: this.groupLength,
          decimal: this.decimal,
          symbol: a ? this.negative : this.positive,
        });
      },
    },
    currency: { symbol: "$" },
    language: { ok: "ok", cancel: "cancel", signin: "signin", signup: "signup" },
  });
  baidu.i18n.currentLocale = "en-US";
  baidu.i18n.cultures["zh-CN"] = baidu.object.extend(baidu.i18n.cultures["zh-CN"] || {}, {
    calendar: {
      dateFormat: "yyyy-MM-dd",
      titleNames: "#{yyyy}年&nbsp;#{MM}月",
      monthNames: ["一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "十二"],
      monthNamesShort: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
      dayNames: { mon: "一", tue: "二", wed: "三", thu: "四", fri: "五", sat: "六", sun: "日" },
    },
    timeZone: 8,
    whitespace: new RegExp("(^[\\s\\t\\xa0\\u3000]+)|([\\u3000\\xa0\\s\\t]+\x24)", "g"),
    number: {
      group: ",",
      groupLength: 3,
      decimal: ".",
      positive: "",
      negative: "-",
      _format: function (b, a) {
        return baidu.i18n.number._format(b, {
          group: this.group,
          groupLength: this.groupLength,
          decimal: this.decimal,
          symbol: a ? this.negative : this.positive,
        });
      },
    },
    currency: { symbol: "￥" },
    language: { ok: "确定", cancel: "取消", signin: "注册", signup: "登录" },
  });
  baidu.i18n.currentLocale = "zh-CN";
  baidu.i18n.number = baidu.i18n.number || {
    format: function (e, c, g) {
      var d = this,
        f = c && baidu.i18n.cultures[c].number,
        b = baidu.i18n.cultures[g || baidu.i18n.currentLocale].number,
        a = false;
      if (typeof e === "string") {
        if (e.indexOf(f.negative) > -1) {
          a = true;
          e = e.replace(f.negative, "");
        } else {
          if (e.indexOf(f.positive) > -1) {
            e = e.replace(f.positive, "");
          }
        }
        e = e.replace(new RegExp(f.group, "g"), "");
      } else {
        if (e < 0) {
          a = true;
          e *= -1;
        }
      }
      e = parseFloat(e);
      if (isNaN(e)) {
        return "NAN";
      }
      return b._format
        ? b._format(e, a)
        : d._format(e, {
            group: b.group || ",",
            decimal: b.decimal || ".",
            groupLength: b.groupLength,
            symbol: a ? b.negative : b.positive,
          });
    },
    _format: function (c, h) {
      var a = String(c).split(h.decimal),
        f = a[0].split("").reverse(),
        b = a[1] || "",
        e = 0,
        g = 0,
        j = "";
      e = parseInt(f.length / h.groupLength);
      g = f.length % h.groupLength;
      e = g == 0 ? e - 1 : e;
      for (var d = 1; d <= e; d++) {
        f.splice(h.groupLength * d + (d - 1), 0, h.group);
      }
      f = f.reverse();
      j = h.symbol + f.join("") + (b.length > 0 ? h.decimal + b : "");
      return j;
    },
  };
  baidu.i18n.currency = baidu.i18n.currency || {
    format: function (e, c, g) {
      var d = this,
        f = c && baidu.i18n.cultures[c].currency,
        b = baidu.i18n.cultures[g || baidu.i18n.currentLocale].currency,
        a;
      if (typeof e === "string") {
        e = e.replace(f.symbol);
      }
      return b.symbol + this._format(e, c, g);
    },
    _format: function (b, a, c) {
      return baidu.i18n.number.format(b, a, c || baidu.i18n.currentLocale);
    },
  };
  baidu.i18n.date = baidu.i18n.date || {
    getDaysInMonth: function (a, b) {
      var c = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
      if (b == 1 && baidu.i18n.date.isLeapYear(a)) {
        return 29;
      }
      return c[b];
    },
    isLeapYear: function (a) {
      return !(a % 400) || (!(a % 4) && !!(a % 100));
    },
    toLocaleDate: function (b, a, c) {
      return this._basicDate(b, a, c || baidu.i18n.currentLocale);
    },
    _basicDate: function (f, c, h) {
      var a = baidu.i18n.cultures[h || baidu.i18n.currentLocale].timeZone,
        g = a * 60,
        b,
        d,
        e = f.getTime();
      if (c) {
        b = baidu.i18n.cultures[c].timeZone;
        d = b * 60;
      } else {
        d = -1 * f.getTimezoneOffset();
        b = d / 60;
      }
      return new Date(b != a ? e + (g - d) * 60000 : e);
    },
    format: function (a, b) {
      var d = baidu.i18n.cultrues[b || baidu.i18n.currentLocale];
      return baidu.date.format(baidu.i18n.date.toLocaleDate(a, "", b), d.calendar.dateFormat);
    },
  };
  baidu.i18n.string = baidu.i18n.string || {
    trim: function (c, a) {
      var b = baidu.i18n.cultures[a || baidu.i18n.currentLocale].whitespace;
      return String(c).replace(b, "");
    },
    translation: function (c, a) {
      var b = baidu.i18n.cultures[a || baidu.i18n.currentLocale].language;
      return b[c] || "";
    },
  };
  baidu.json.decode = baidu.json.parse;
  baidu.json.stringify = (function () {
    var b = { "\b": "\\b", "\t": "\\t", "\n": "\\n", "\f": "\\f", "\r": "\\r", '"': '\\"', "\\": "\\\\" };
    function a(f) {
      if (/["\\\x00-\x1f]/.test(f)) {
        f = f.replace(/["\\\x00-\x1f]/g, function (g) {
          var h = b[g];
          if (h) {
            return h;
          }
          h = g.charCodeAt();
          return "\\u00" + Math.floor(h / 16).toString(16) + (h % 16).toString(16);
        });
      }
      return '"' + f + '"';
    }
    function d(m) {
      var g = ["["],
        h = m.length,
        f,
        j,
        k;
      for (j = 0; j < h; j++) {
        k = m[j];
        switch (typeof k) {
          case "undefined":
          case "function":
          case "unknown":
            break;
          default:
            if (f) {
              g.push(",");
            }
            g.push(baidu.json.stringify(k));
            f = 1;
        }
      }
      g.push("]");
      return g.join("");
    }
    function c(f) {
      return f < 10 ? "0" + f : f;
    }
    function e(f) {
      return (
        '"' +
        f.getFullYear() +
        "-" +
        c(f.getMonth() + 1) +
        "-" +
        c(f.getDate()) +
        "T" +
        c(f.getHours()) +
        ":" +
        c(f.getMinutes()) +
        ":" +
        c(f.getSeconds()) +
        '"'
      );
    }
    return function (k) {
      switch (typeof k) {
        case "undefined":
          return "undefined";
        case "number":
          return isFinite(k) ? String(k) : "null";
        case "string":
          return a(k);
        case "boolean":
          return String(k);
        default:
          if (k === null) {
            return "null";
          } else {
            if (baidu.type(k) === "array") {
              return d(k);
            } else {
              if (baidu.type(k) === "date") {
                return e(k);
              } else {
                var g = ["{"],
                  j = baidu.json.stringify,
                  f,
                  i;
                for (var h in k) {
                  if (Object.prototype.hasOwnProperty.call(k, h)) {
                    i = k[h];
                    switch (typeof i) {
                      case "undefined":
                      case "unknown":
                      case "function":
                        break;
                      default:
                        if (f) {
                          g.push(",");
                        }
                        f = 1;
                        g.push(j(h) + ":" + j(i));
                    }
                  }
                }
                g.push("}");
                return g.join("");
              }
            }
          }
      }
    };
  })();
  baidu.json.encode = baidu.json.stringify;
  baidu.lang.Class.prototype.addEventListeners = function (c, d) {
    if (typeof d == "undefined") {
      for (var b in c) {
        this.addEventListener(b, c[b]);
      }
    } else {
      c = c.split(",");
      var b = 0,
        a = c.length,
        e;
      for (; b < a; b++) {
        this.addEventListener(baidu.trim(c[b]), d);
      }
    }
  };
  baidu.lang.createClass = function (b, j) {
    j = j || {};
    var g = j.superClass || baidu.lang.Class;
    var h = function () {
      var m = this;
      j.decontrolled && (m.__decontrolled = true);
      g.apply(m, arguments);
      for (k in h.options) {
        m[k] = h.options[k];
      }
      b.apply(m, arguments);
      for (var k = 0, l = h["\x06r"]; l && k < l.length; k++) {
        l[k].apply(m, arguments);
      }
    };
    h.options = j.options || {};
    var a = function () {},
      f = b.prototype;
    a.prototype = g.prototype;
    var d = (h.prototype = new a());
    for (var c in f) {
      d[c] = f[c];
    }
    var e = j.className || j.type;
    typeof e == "string" && (d.__type = e);
    d.constructor = f.constructor;
    h.extend = function (l) {
      for (var k in l) {
        h.prototype[k] = l[k];
      }
      return h;
    };
    return h;
  };
  baidu.lang.decontrol = (function () {
    var a = baidu.global("_maps_id");
    return function (b) {
      delete a[b];
    };
  })();
  baidu.lang.eventCenter = baidu.lang.eventCenter || baidu.lang.createSingle();
  baidu.lang.getModule = function (b, c) {
    var d = b.split("."),
      e = c || window,
      a;
    for (; (a = d.shift());) {
      if (e[a] != null) {
        e = e[a];
      } else {
        return null;
      }
    }
    return e;
  };
  baidu.lang.isBoolean = baidu.isBoolean;
  baidu.lang.isDate = baidu.isDate;
  baidu.lang.isElement = baidu.isElement;
  baidu.lang.isObject = baidu.isObject;
  baidu.lang.isWindow = function (a) {
    return baidu.type(a) == "Window";
  };
  baidu.lang.module = function (d, g, b) {
    var h = d.split("."),
      a = h.length - 1,
      c,
      f = 0;
    if (!b) {
      try {
        if (!new RegExp("^[a-zA-Z_\x24][a-zA-Z0-9_\x24]*\x24").test(h[0])) {
          throw "";
        }
        b = window["eval"](h[0]);
        f = 1;
      } catch (j) {
        b = window;
      }
    }
    for (; f < a; f++) {
      c = h[f];
      if (!b[c]) {
        b[c] = {};
      }
      b = b[c];
    }
    if (!b[h[a]]) {
      b[h[a]] = g;
    }
  };
  baidu.lang.register = function (b, d, a) {
    var c = b["\x06r"] || (b["\x06r"] = []);
    c[c.length] = d;
    for (var e in a) {
      b.prototype[e] = a[e];
    }
  };
  baidu.global("_maps_id");
  baidu.number.extend({
    comma: function (a) {
      var b = this;
      if (!a || a < 1) {
        a = 3;
      }
      b = String(b).split(".");
      b[0] = b[0].replace(new RegExp("(\\d)(?=(\\d{" + a + "})+$)", "ig"), "$1,");
      return b.join(".");
    },
  });
  baidu.number.randomInt = function (b, a) {
    return Math.floor(Math.random() * (a - b + 1) + b);
  };
  baidu.object.clone = function (e) {
    var b = e,
      c,
      a;
    if (!e || e instanceof Number || e instanceof String || e instanceof Boolean) {
      return b;
    } else {
      if (baidu.lang.isArray(e)) {
        b = [];
        var d = 0;
        for (c = 0, a = e.length; c < a; c++) {
          b[d++] = baidu.object.clone(e[c]);
        }
      } else {
        if (baidu.object.isPlain(e)) {
          b = {};
          for (c in e) {
            if (e.hasOwnProperty(c)) {
              b[c] = baidu.object.clone(e[c]);
            }
          }
        }
      }
    }
    return b;
  };
  baidu.object.keys = function (d) {
    var a = [],
      c = 0,
      b;
    for (b in d) {
      if (d.hasOwnProperty(b)) {
        a[c++] = b;
      }
    }
    return a;
  };
  baidu.object.map = function (d, c) {
    var b = {};
    for (var a in d) {
      if (d.hasOwnProperty(a)) {
        b[a] = c(d[a], a);
      }
    }
    return b;
  };
  (function () {
    var b = function (c) {
      return baidu.lang.isObject(c) && !baidu.lang.isFunction(c);
    };
    function a(g, f, e, d, c) {
      if (f.hasOwnProperty(e)) {
        if (c && b(g[e])) {
          baidu.object.merge(g[e], f[e], { overwrite: d, recursive: c });
        } else {
          if (d || !(e in g)) {
            g[e] = f[e];
          }
        }
      }
    }
    baidu.object.merge = function (h, c, k) {
      var e = 0,
        l = k || {},
        g = l["overwrite"],
        j = l["whiteList"],
        d = l["recursive"],
        f;
      if (j && j.length) {
        f = j.length;
        for (; e < f; ++e) {
          a(h, c, j[e], g, d);
        }
      } else {
        for (e in c) {
          a(h, c, e, g, d);
        }
      }
      return h;
    };
  })();
  baidu.page.createStyleSheet = function (a) {
    var f = a || {},
      d = f.document || document,
      c;
    if (baidu.browser.ie) {
      if (!f.url) {
        f.url = "";
      }
      return d.createStyleSheet(f.url, f.index);
    } else {
      c = "<style type='text/css'></style>";
      f.url && (c = "<link type='text/css' rel='stylesheet' href='" + f.url + "'/>");
      baidu.dom.insertHTML(d.getElementsByTagName("HEAD")[0], "beforeEnd", c);
      if (f.url) {
        return null;
      }
      var b = d.styleSheets[d.styleSheets.length - 1],
        e = b.rules || b.cssRules;
      return {
        self: b,
        rules: b.rules || b.cssRules,
        addRule: function (g, j, h) {
          if (b.addRule) {
            return b.addRule(g, j, h);
          } else {
            if (b.insertRule) {
              isNaN(h) && (h = e.length);
              return b.insertRule(g + "{" + j + "}", h);
            }
          }
        },
        removeRule: function (g) {
          if (b.removeRule) {
            b.removeRule(g);
          } else {
            if (b.deleteRule) {
              isNaN(g) && (g = 0);
              b.deleteRule(g);
            }
          }
        },
      };
    }
  };
  baidu.page.getHeight = function () {
    var d = document,
      a = d.body,
      c = d.documentElement,
      b = d.compatMode == "BackCompat" ? a : d.documentElement;
    return Math.max(c.scrollHeight, a.scrollHeight, b.clientHeight);
  };
  baidu.page.getViewHeight = function () {
    var b = document,
      a = b.compatMode == "BackCompat" ? b.body : b.documentElement;
    return a.clientHeight;
  };
  baidu.page.getViewWidth = function () {
    var b = document,
      a = b.compatMode == "BackCompat" ? b.body : b.documentElement;
    return a.clientWidth;
  };
  baidu.page.getWidth = function () {
    var d = document,
      a = d.body,
      c = d.documentElement,
      b = d.compatMode == "BackCompat" ? a : d.documentElement;
    return Math.max(c.scrollWidth, a.scrollWidth, b.clientWidth);
  };
  baidu.page.lazyLoadImage = function (a) {
    a = a || {};
    a.preloadHeight = a.preloadHeight || 0;
    baidu.dom.ready(function () {
      var e = document.getElementsByTagName("IMG"),
        f = e,
        g = e.length,
        d = 0,
        k = c(),
        j = "data-tangram-ori-src",
        h;
      if (a.className) {
        f = [];
        for (; d < g; ++d) {
          if (baidu.dom.hasClass(e[d], a.className)) {
            f.push(e[d]);
          }
        }
      }
      function c() {
        return baidu.page.getScrollTop() + baidu.page.getViewHeight() + a.preloadHeight;
      }
      for (d = 0, g = f.length; d < g; ++d) {
        h = f[d];
        if (baidu.dom.getPosition(h).top > k) {
          h.setAttribute(j, h.src);
          a.placeHolder ? (h.src = a.placeHolder) : h.removeAttribute("src");
        }
      }
      var b = function () {
        var m = c(),
          o,
          p = true,
          n = 0,
          l = f.length;
        for (; n < l; ++n) {
          h = f[n];
          o = h.getAttribute(j);
          o && (p = false);
          if (baidu.dom.getPosition(h).top < m && o) {
            h.src = o;
            h.removeAttribute(j);
            baidu.lang.isFunction(a.onlazyload) && a.onlazyload(h);
          }
        }
        p && baidu.dom(window).off("scroll", b);
      };
      baidu.dom(window).on("scroll", b);
    });
  };
  baidu.page.load = function (c, j, e) {
    j = j || {};
    var h = baidu.page.load,
      a = (h._cache = h._cache || {}),
      g = (h._loadingCache = h._loadingCache || {}),
      f = j.parallel;
    function d() {
      for (var l = 0, k = c.length; l < k; ++l) {
        if (!a[c[l].url]) {
          setTimeout(arguments.callee, 10);
          return;
        }
      }
      j.onload();
    }
    function b(m, p) {
      var o, l, k;
      switch (m.type.toLowerCase()) {
        case "css":
          o = document.createElement("link");
          o.setAttribute("rel", "stylesheet");
          o.setAttribute("type", "text/css");
          break;
        case "js":
          o = document.createElement("script");
          o.setAttribute("type", "text/javascript");
          o.setAttribute("charset", m.charset || h.charset);
          break;
        case "html":
          o = document.createElement("iframe");
          o.frameBorder = "none";
          break;
        default:
          return;
      }
      var n = baidu.dom(o);
      k = function () {
        if (!l && (!this.readyState || this.readyState === "loaded" || this.readyState === "complete")) {
          l = true;
          n.off("load", k);
          n.off("readystatechange", k);
          p.call(window, o);
        }
      };
      n.on("load", k);
      n.on("readystatechange", k);
      if (m.type == "css") {
        (function () {
          if (l) {
            return;
          }
          try {
            o.sheet.cssRule;
          } catch (q) {
            setTimeout(arguments.callee, 20);
            return;
          }
          l = true;
          p.call(window, o);
        })();
      }
      o.href = o.src = m.url;
      document.getElementsByTagName("head")[0].appendChild(o);
    }
    baidu.lang.isString(c) && (c = [{ url: c }]);
    if (!(c && c.length)) {
      return;
    }
    function i(m) {
      var l = m.url,
        n = !!f,
        k,
        o = function (p) {
          a[m.url] = p;
          delete g[m.url];
          if (baidu.lang.isFunction(m.onload)) {
            if (false === m.onload.call(window, p)) {
              return;
            }
          }
          !f && h(c.slice(1), j, true);
          if (!e && baidu.lang.isFunction(j.onload)) {
            d();
          }
        };
      m.type = m.type || l.replace(/^[^\?#]+\.(css|js|html)(\?|#| |$)[^\?#]*/i, "$1");
      m.requestType = m.requestType || (m.type == "html" ? "ajax" : "dom");
      if ((k = a[m.url])) {
        o(k);
        return n;
      }
      if (!j.refresh && g[m.url]) {
        setTimeout(function () {
          i(m);
        }, 10);
        return n;
      }
      g[m.url] = true;
      if (m.requestType.toLowerCase() == "dom") {
        b(m, o);
      } else {
        baidu.ajax.get(m.url, function (q, p) {
          o(p);
        });
      }
      return n;
    }
    baidu.forEach(c, i);
  };
  baidu.page.load.charset = "UTF8";
  baidu.page.loadCssFile = function (b) {
    var a = document.createElement("link");
    a.setAttribute("rel", "stylesheet");
    a.setAttribute("type", "text/css");
    a.setAttribute("href", b);
    document.getElementsByTagName("head")[0].appendChild(a);
  };
  baidu.page.loadJsFile = function (b) {
    var a = document.createElement("script");
    a.setAttribute("type", "text/javascript");
    a.setAttribute("src", b);
    a.setAttribute("defer", "defer");
    document.getElementsByTagName("head")[0].appendChild(a);
  };
  baidu.param = function (a) {
    a = a || arguments.callee.caller.arguments;
    var c = "",
      d = a.length;
    for (var b = 0; b < d; b++) {
      c += "," + baidu.type(a[b]);
    }
    return c ? c.substr(1) : "";
  };
  baidu.platform =
    baidu.platform ||
    (function () {
      var b = navigator.userAgent,
        a = function () {};
      baidu.forEach("Android iPad iPhone Linux Macintosh Windows X11".split(" "), function (d) {
        var c = d.charAt(0).toUpperCase() + d.toLowerCase().substr(1);
        baidu["is" + c] = a["is" + c] = b.indexOf(d) > -1;
      });
      return a;
    })();
  baidu.post = baidu.post || baidu._util_.smartAjax("post");
  baidu.regexp = baidu.regexp || {};
  baidu.regexp.compile = (function (a) {
    return function (b, e) {
      var d,
        c = b + "$$" + e;
      !(d = a[c]) && (d = a[c] = new RegExp(b, e));
      d.lastIndex > 0 && (d.lastIndex = 0);
      return d;
    };
  })(baidu.global("_maps_RegExp"));
  baidu.createChain(
    "sio",
    function (a) {
      switch (typeof a) {
        case "string":
          return new baidu.sio.$Sio(a);
          break;
      }
    },
    function (a) {
      this.url = a;
    },
  );
  baidu.sio._createScriptTag = function (b, a, c) {
    b.setAttribute("type", "text/javascript");
    c && b.setAttribute("charset", c);
    b.setAttribute("src", a);
    document.getElementsByTagName("head")[0].appendChild(b);
  };
  baidu.sio._removeScriptTag = function (b) {
    if (b.clearAttributes) {
      b.clearAttributes();
    } else {
      for (var a in b) {
        if (b.hasOwnProperty(a)) {
          delete b[a];
        }
      }
    }
    if (b && b.parentNode) {
      b.parentNode.removeChild(b);
    }
    b = null;
  };
  baidu.sio.extend({
    callByBrowser: function (g, i) {
      var a = this.url;
      var d = document.createElement("SCRIPT"),
        e = 0,
        j = i || {},
        c = j["charset"],
        h = g || function () {},
        f = j["timeOut"] || 0,
        b;
      d.onload = d.onreadystatechange = function () {
        if (e) {
          return;
        }
        var k = d.readyState;
        if ("undefined" == typeof k || k == "loaded" || k == "complete") {
          e = 1;
          try {
            h();
            clearTimeout(b);
          } finally {
            d.onload = d.onreadystatechange = null;
            baidu.sio._removeScriptTag(d);
          }
        }
      };
      if (f) {
        b = setTimeout(function () {
          d.onload = d.onreadystatechange = null;
          baidu.sio._removeScriptTag(d);
          j.onfailure && j.onfailure();
        }, f);
      }
      baidu.sio._createScriptTag(d, a, c);
    },
  });
  baidu.sio.extend({
    callByServer: function (m, n) {
      var a = this.url;
      var i = document.createElement("SCRIPT"),
        h = "bd__cbs__",
        k,
        e,
        o = n || {},
        d = o["charset"],
        f = o["queryField"] || "callback",
        l = o["timeOut"] || 0,
        b,
        c = new RegExp("(\\?|&)" + f + "=([^&]*)"),
        g;
      if (baidu.lang.isFunction(m)) {
        k = h + Math.floor(Math.random() * 2147483648).toString(36);
        window[k] = j(0);
      } else {
        if (baidu.lang.isString(m)) {
          k = m;
        } else {
          if ((g = c.exec(a))) {
            k = g[2];
          }
        }
      }
      if (l) {
        b = setTimeout(j(1), l);
      }
      a = a.replace(c, "\x241" + f + "=" + k);
      if (a.search(c) < 0) {
        a += (a.indexOf("?") < 0 ? "?" : "&") + f + "=" + k;
      }
      baidu.sio._createScriptTag(i, a, d);
      function j(p) {
        return function () {
          try {
            if (p) {
              o.onfailure && o.onfailure();
            } else {
              m.apply(window, arguments);
              clearTimeout(b);
            }
            window[k] = null;
            delete window[k];
          } catch (q) {
          } finally {
            baidu.sio._removeScriptTag(i);
          }
        };
      }
    },
  });
  baidu.sio.extend({
    log: function () {
      url = this.url;
      var a = new Image(),
        b = "tangram_sio_log_" + Math.floor(Math.random() * 2147483648).toString(36);
      window[b] = a;
      a.onload =
        a.onerror =
        a.onabort =
          function () {
            a.onload = a.onerror = a.onabort = null;
            window[b] = null;
            a = null;
          };
      a.src = url;
    },
  });
  (function (M, q) {
    baidu.query = function (Q, S, R) {
      return baidu.merge(R || [], baidu.sizzle(Q, S));
    };
    var e = M.document,
      j = e.documentElement,
      O = "sizcache" + (Math.random() + "").replace(".", ""),
      o = 0,
      d = Object.prototype.toString,
      A = "undefined",
      l = false,
      h = true,
      H = /^#([\w\-]+$)|^(\w+$)|^\.([\w\-]+$)/,
      z =
        /((?:\((?:\([^()]+\)|[^()]+)+\)|\[(?:\[[^\[\]]*\]|['"][^'"]*['"]|[^\[\]'"]+)+\]|\\.|[^ >+~,(\[\\]+)+|[>+~])(\s*,\s*)?((?:.|\r|\n)*)/g,
      y = /\\/g,
      P = /\W/,
      g = /^\w/,
      s = /\D/,
      f = /(-?)(\d*)(?:n([+\-]?\d*))?/,
      w = /^\+|\s*/g,
      v = /h\d/i,
      I = /input|select|textarea|button/i,
      m = /[\t\n\f\r]/g,
      r = "(?:[-\\w]|[^\\x00-\\xa0]|\\\\.)",
      E = {
        ID: new RegExp("#(" + r + "+)"),
        CLASS: new RegExp("\\.(" + r + "+)"),
        NAME: new RegExp("\\[name=['\"]*(" + r + "+)['\"]*\\]"),
        TAG: new RegExp("^(" + r.replace("[-", "[-\\*") + "+)"),
        ATTR: new RegExp("\\[\\s*(" + r + "+)\\s*(?:(\\S?=)\\s*(?:(['\"])(.*?)\\3|(#?" + r + "*)|)|)\\s*\\]"),
        PSEUDO: new RegExp(":(" + r + "+)(?:\\((['\"]?)((?:\\([^\\)]+\\)|[^\\(\\)]*)+)\\2\\))?"),
        CHILD: /:(only|nth|last|first)-child(?:\(\s*(even|odd|(?:[+\-]?\d+|(?:[+\-]?\d*)?n\s*(?:[+\-]\s*\d+)?))\s*\))?/,
        POS: /:(nth|eq|gt|lt|first|last|even|odd)(?:\((\d*)\))?(?=[^\-]|$)/,
      },
      J = E.POS,
      K = (function () {
        var R,
          S = function (V, U) {
            return "\\" + (U - 0 + 1);
          },
          Q = {};
        for (R in E) {
          E[R] = new RegExp(E[R].source + /(?![^\[]*\])(?![^\(]*\))/.source);
          Q[R] = new RegExp(/(^(?:.|\r|\n)*?)/.source + E[R].source.replace(/\\(\d+)/g, S));
        }
        E.globalPOS = J;
        return Q;
      })(),
      G = function (Q) {
        var R = false,
          U = e.createElement("div");
        try {
          R = Q(U);
        } catch (S) {}
        U = null;
        return R;
      },
      i = G(function (S) {
        var Q = true,
          R = "script" + new Date().getTime();
        S.innerHTML = "<a name ='" + R + "'/>";
        j.insertBefore(S, j.firstChild);
        if (e.getElementById(R)) {
          Q = false;
        }
        j.removeChild(S);
        return Q;
      }),
      c = G(function (Q) {
        Q.appendChild(e.createComment(""));
        return Q.getElementsByTagName("*").length === 0;
      }),
      D = G(function (Q) {
        Q.innerHTML = "<a href='#'></a>";
        return Q.firstChild && typeof Q.firstChild.getAttribute !== A && Q.firstChild.getAttribute("href") === "#";
      }),
      C = G(function (Q) {
        Q.innerHTML = "<div class='test e'></div><div class='test'></div>";
        if (!Q.getElementsByClassName || Q.getElementsByClassName("e").length === 0) {
          return false;
        }
        Q.lastChild.className = "e";
        return Q.getElementsByClassName("e").length !== 1;
      });
    [0, 0].sort(function () {
      h = false;
      return 0;
    });
    var L = function (Q, W, V) {
      V = V || [];
      W = W || e;
      var S,
        X,
        U,
        R = W.nodeType;
      if (R !== 1 && R !== 9) {
        return [];
      }
      if (!Q || typeof Q !== "string") {
        return V;
      }
      U = u(W);
      if (!U) {
        if ((S = H.exec(Q))) {
          if (S[1]) {
            if (R === 9) {
              X = W.getElementById(S[1]);
              if (X && X.parentNode) {
                if (X.id === S[1]) {
                  return t([X], V);
                }
              } else {
                return t([], V);
              }
            } else {
              if (W.ownerDocument && (X = W.ownerDocument.getElementById(S[1])) && B(W, X) && X.id === S[1]) {
                return t([X], V);
              }
            }
          } else {
            if (S[2]) {
              if (Q === "body" && W.body) {
                return t([W.body], V);
              }
              return t(W.getElementsByTagName(Q), V);
            } else {
              if (C && S[3] && W.getElementsByClassName) {
                return t(W.getElementsByClassName(S[3]), V);
              }
            }
          }
        }
      }
      return N(Q, W, V, q, U);
    };
    var N = function (W, Q, Z, aa, V) {
      var S,
        ae,
        ah,
        R,
        ad,
        ag,
        af,
        Y,
        ac = Q,
        U = true,
        X = [],
        ab = W;
      do {
        z.exec("");
        S = z.exec(ab);
        if (S) {
          ab = S[3];
          X.push(S[1]);
          if (S[2]) {
            R = S[3];
            break;
          }
        }
      } while (S);
      if (X.length > 1 && J.exec(W)) {
        if (X.length === 2 && F.relative[X[0]]) {
          ae = x(X[0] + X[1], Q, aa, V);
        } else {
          ae = F.relative[X[0]] ? [Q] : L(X.shift(), Q);
          while (X.length) {
            W = X.shift();
            if (F.relative[W]) {
              W += X.shift();
            }
            ae = x(W, ae, aa, V);
          }
        }
      } else {
        if (!aa && X.length > 1 && Q.nodeType === 9 && !V && E.ID.test(X[0]) && !E.ID.test(X[X.length - 1])) {
          ad = L.find(X.shift(), Q, V);
          Q = ad.expr ? L.filter(ad.expr, ad.set)[0] : ad.set[0];
        }
        if (Q) {
          ad = aa
            ? { expr: X.pop(), set: t(aa) }
            : L.find(X.pop(), (X.length >= 1 && (X[0] === "~" || X[0] === "+") && Q.parentNode) || Q, V);
          ae = ad.expr ? L.filter(ad.expr, ad.set) : ad.set;
          if (X.length > 0) {
            ah = t(ae);
          } else {
            U = false;
          }
          while (X.length) {
            ag = X.pop();
            af = ag;
            if (!F.relative[ag]) {
              ag = "";
            } else {
              af = X.pop();
            }
            if (af == null) {
              af = Q;
            }
            F.relative[ag](ah, af, V);
          }
        } else {
          ah = X = [];
        }
      }
      if (!ah) {
        ah = ae;
      }
      if (!ah) {
        L.error(ag || W);
      }
      if (d.call(ah) === "[object Array]") {
        if (!U) {
          Z.push.apply(Z, ah);
        } else {
          if (Q && Q.nodeType === 1) {
            for (Y = 0; ah[Y] != null; Y++) {
              if (ah[Y] && (ah[Y] === true || (ah[Y].nodeType === 1 && B(Q, ah[Y])))) {
                Z.push(ae[Y]);
              }
            }
          } else {
            for (Y = 0; ah[Y] != null; Y++) {
              if (ah[Y] && ah[Y].nodeType === 1) {
                Z.push(ae[Y]);
              }
            }
          }
        }
      } else {
        t(ah, Z);
      }
      if (R) {
        N(R, ac, Z, aa, V);
        n(Z);
      }
      return Z;
    };
    var u = (L.isXML = function (Q) {
      var R = (Q ? Q.ownerDocument || Q : 0).documentElement;
      return R ? R.nodeName !== "HTML" : false;
    });
    var t = function (U, S) {
      S = S || [];
      var R = 0,
        Q = U.length;
      if (typeof Q === "number") {
        for (; R < Q; R++) {
          S.push(U[R]);
        }
      } else {
        for (; U[R]; R++) {
          S.push(U[R]);
        }
      }
      return S;
    };
    var n = (L.uniqueSort = function (R) {
      if (p) {
        l = h;
        R.sort(p);
        if (l) {
          for (var Q = 1; Q < R.length; Q++) {
            if (R[Q] === R[Q - 1]) {
              R.splice(Q--, 1);
            }
          }
        }
      }
      return R;
    });
    var B = (L.contains = j.compareDocumentPosition
      ? function (R, Q) {
          return !!(R.compareDocumentPosition(Q) & 16);
        }
      : j.contains
        ? function (R, Q) {
            return R !== Q && (R.contains ? R.contains(Q) : false);
          }
        : function (R, Q) {
            while ((Q = Q.parentNode)) {
              if (Q === R) {
                return true;
              }
            }
            return false;
          });
    L.matches = function (Q, R) {
      return N(Q, e, [], R, u(e));
    };
    L.matchesSelector = function (Q, R) {
      return N(R, e, [], [Q], u(e)).length > 0;
    };
    L.find = function (Z, Q, S) {
      var Y, U, W, V, X, R;
      if (!Z) {
        return [];
      }
      for (U = 0, W = F.order.length; U < W; U++) {
        X = F.order[U];
        if ((V = K[X].exec(Z))) {
          R = V[1];
          V.splice(1, 1);
          if (R.substr(R.length - 1) !== "\\") {
            V[1] = (V[1] || "").replace(y, "");
            Y = F.find[X](V, Q, S);
            if (Y != null) {
              Z = Z.replace(E[X], "");
              break;
            }
          }
        }
      }
      if (!Y) {
        Y = typeof Q.getElementsByTagName !== A ? Q.getElementsByTagName("*") : [];
      }
      return { set: Y, expr: Z };
    };
    L.filter = function (ac, ab, af, V) {
      var X,
        Q,
        aa,
        ah,
        ae,
        R,
        U,
        W,
        ad,
        S = ac,
        ag = [],
        Z = ab,
        Y = ab && ab[0] && u(ab[0]);
      while (ac && ab.length) {
        for (aa in F.filter) {
          if ((X = K[aa].exec(ac)) != null && X[2]) {
            R = F.filter[aa];
            U = X[1];
            Q = false;
            X.splice(1, 1);
            if (U.substr(U.length - 1) === "\\") {
              continue;
            }
            if (Z === ag) {
              ag = [];
            }
            if (F.preFilter[aa]) {
              X = F.preFilter[aa](X, Z, af, ag, V, Y);
              if (!X) {
                Q = ah = true;
              } else {
                if (X === true) {
                  continue;
                }
              }
            }
            if (X) {
              for (W = 0; (ae = Z[W]) != null; W++) {
                if (ae) {
                  ah = R(ae, X, W, Z);
                  ad = V ^ ah;
                  if (af && ah != null) {
                    if (ad) {
                      Q = true;
                    } else {
                      Z[W] = false;
                    }
                  } else {
                    if (ad) {
                      ag.push(ae);
                      Q = true;
                    }
                  }
                }
              }
            }
            if (ah !== q) {
              if (!af) {
                Z = ag;
              }
              ac = ac.replace(E[aa], "");
              if (!Q) {
                return [];
              }
              break;
            }
          }
        }
        if (ac === S) {
          if (Q == null) {
            L.error(ac);
          } else {
            break;
          }
        }
        S = ac;
      }
      return Z;
    };
    L.error = function (Q) {
      throw new Error("Syntax error, unrecognized expression: " + Q);
    };
    var a = (L.getText = function (V) {
      var S,
        U,
        Q = V.nodeType,
        R = "";
      if (Q) {
        if (Q === 1 || Q === 9 || Q === 11) {
          if (typeof V.textContent === "string") {
            return V.textContent;
          } else {
            for (V = V.firstChild; V; V = V.nextSibling) {
              R += a(V);
            }
          }
        } else {
          if (Q === 3 || Q === 4) {
            return V.nodeValue;
          }
        }
      } else {
        for (S = 0; (U = V[S]); S++) {
          if (U.nodeType !== 8) {
            R += a(U);
          }
        }
      }
      return R;
    });
    var F = (L.selectors = {
      match: E,
      leftMatch: K,
      order: ["ID", "NAME", "TAG"],
      attrMap: { class: "className", for: "htmlFor" },
      attrHandle: {
        href: D
          ? function (Q) {
              return Q.getAttribute("href");
            }
          : function (Q) {
              return Q.getAttribute("href", 2);
            },
        type: function (Q) {
          return Q.getAttribute("type");
        },
      },
      relative: {
        "+": function (X, R) {
          var U = typeof R === "string",
            W = U && !P.test(R),
            Y = U && !W;
          if (W) {
            R = R.toLowerCase();
          }
          for (var S = 0, Q = X.length, V; S < Q; S++) {
            if ((V = X[S])) {
              while ((V = V.previousSibling) && V.nodeType !== 1) {}
              X[S] = Y || (V && V.nodeName.toLowerCase() === R) ? V || false : V === R;
            }
          }
          if (Y) {
            L.filter(R, X, true);
          }
        },
        ">": function (X, R) {
          var W,
            V = typeof R === "string",
            S = 0,
            Q = X.length;
          if (V && !P.test(R)) {
            R = R.toLowerCase();
            for (; S < Q; S++) {
              W = X[S];
              if (W) {
                var U = W.parentNode;
                X[S] = U.nodeName.toLowerCase() === R ? U : false;
              }
            }
          } else {
            for (; S < Q; S++) {
              W = X[S];
              if (W) {
                X[S] = V ? W.parentNode : W.parentNode === R;
              }
            }
            if (V) {
              L.filter(R, X, true);
            }
          }
        },
        "": function (S, R, Q) {
          k("parentNode", S, R, Q);
        },
        "~": function (S, R, Q) {
          k("previousSibling", S, R, Q);
        },
      },
      find: {
        ID: i
          ? function (S, U, R) {
              if (typeof U.getElementById !== A && !R) {
                var Q = U.getElementById(S[1]);
                return Q && Q.parentNode ? [Q] : [];
              }
            }
          : function (S, U, R) {
              if (typeof U.getElementById !== A && !R) {
                var Q = U.getElementById(S[1]);
                return Q
                  ? Q.id === S[1] || (typeof Q.getAttributeNode !== A && Q.getAttributeNode("id").nodeValue === S[1])
                    ? [Q]
                    : q
                  : [];
              }
            },
        NAME: function (S, W) {
          if (typeof W.getElementsByName !== A) {
            var R = [],
              V = W.getElementsByName(S[1]),
              U = 0,
              Q = V.length;
            for (; U < Q; U++) {
              if (V[U].getAttribute("name") === S[1]) {
                R.push(V[U]);
              }
            }
            return R.length === 0 ? null : R;
          }
        },
        TAG: c
          ? function (Q, R) {
              if (typeof R.getElementsByTagName !== A) {
                return R.getElementsByTagName(Q[1]);
              }
            }
          : function (Q, V) {
              var U = V.getElementsByTagName(Q[1]);
              if (Q[1] === "*") {
                var S = [],
                  R = 0;
                for (; U[R]; R++) {
                  if (U[R].nodeType === 1) {
                    S.push(U[R]);
                  }
                }
                U = S;
              }
              return U;
            },
      },
      preFilter: {
        CLASS: function (V, R, S, Q, Y, U) {
          V = " " + V[1].replace(y, "") + " ";
          if (U) {
            return V;
          }
          for (var W = 0, X; (X = R[W]) != null; W++) {
            if (X) {
              if (Y ^ (X.className && (" " + X.className + " ").replace(m, " ").indexOf(V) >= 0)) {
                if (!S) {
                  Q.push(X);
                }
              } else {
                if (S) {
                  R[W] = false;
                }
              }
            }
          }
          return false;
        },
        ID: function (Q) {
          return Q[1].replace(y, "");
        },
        TAG: function (R, Q) {
          return R[1].replace(y, "").toLowerCase();
        },
        CHILD: function (Q) {
          if (Q[1] === "nth") {
            if (!Q[2]) {
              L.error(Q[0]);
            }
            Q[2] = Q[2].replace(w, "");
            var R = f.exec(
              (Q[2] === "even" && "2n") || (Q[2] === "odd" && "2n+1") || (!s.test(Q[2]) && "0n+" + Q[2]) || Q[2],
            );
            Q[2] = R[1] + (R[2] || 1) - 0;
            Q[3] = R[3] - 0;
          } else {
            if (Q[2]) {
              L.error(Q[0]);
            }
          }
          Q[0] = o++;
          return Q;
        },
        ATTR: function (W, R, S, Q, X, V) {
          var U = (W[1] = W[1].replace(y, ""));
          if (!V && F.attrMap[U]) {
            W[1] = F.attrMap[U];
          }
          W[4] = (W[4] || W[5] || "").replace(y, "");
          if (W[2] === "~=") {
            W[4] = " " + W[4] + " ";
          }
          return W;
        },
        PSEUDO: function (W, R, S, Q, X, V) {
          if (W[1] === "not") {
            if ((z.exec(W[3]) || "").length > 1 || g.test(W[3])) {
              W[3] = N(W[3], e, [], R, V);
            } else {
              var U = L.filter(W[3], R, S, !X);
              if (!S) {
                Q.push.apply(Q, U);
              }
              return false;
            }
          } else {
            if (E.POS.test(W[0]) || E.CHILD.test(W[0])) {
              return true;
            }
          }
          return W;
        },
        POS: function (Q) {
          Q.unshift(true);
          return Q;
        },
      },
      filters: {
        enabled: function (Q) {
          return Q.disabled === false;
        },
        disabled: function (Q) {
          return Q.disabled === true;
        },
        checked: function (Q) {
          var R = Q.nodeName.toLowerCase();
          return (R === "input" && !!Q.checked) || (R === "option" && !!Q.selected);
        },
        selected: function (Q) {
          if (Q.parentNode) {
            Q.parentNode.selectedIndex;
          }
          return Q.selected === true;
        },
        parent: function (Q) {
          return !!Q.firstChild;
        },
        empty: function (Q) {
          return !Q.firstChild;
        },
        has: function (S, R, Q) {
          return !!L(Q[3], S).length;
        },
        header: function (Q) {
          return v.test(Q.nodeName);
        },
        text: function (S) {
          var Q = S.getAttribute("type"),
            R = S.type;
          return S.nodeName.toLowerCase() === "input" && "text" === R && (Q === null || Q.toLowerCase() === R);
        },
        radio: function (Q) {
          return Q.nodeName.toLowerCase() === "input" && "radio" === Q.type;
        },
        checkbox: function (Q) {
          return Q.nodeName.toLowerCase() === "input" && "checkbox" === Q.type;
        },
        file: function (Q) {
          return Q.nodeName.toLowerCase() === "input" && "file" === Q.type;
        },
        password: function (Q) {
          return Q.nodeName.toLowerCase() === "input" && "password" === Q.type;
        },
        submit: function (R) {
          var Q = R.nodeName.toLowerCase();
          return (Q === "input" || Q === "button") && "submit" === R.type;
        },
        image: function (Q) {
          return Q.nodeName.toLowerCase() === "input" && "image" === Q.type;
        },
        reset: function (R) {
          var Q = R.nodeName.toLowerCase();
          return (Q === "input" || Q === "button") && "reset" === R.type;
        },
        button: function (R) {
          var Q = R.nodeName.toLowerCase();
          return (Q === "input" && "button" === R.type) || Q === "button";
        },
        input: function (Q) {
          return I.test(Q.nodeName);
        },
        focus: function (Q) {
          var R = Q.ownerDocument;
          return Q === R.activeElement && (!R.hasFocus || R.hasFocus()) && !!(Q.type || Q.href);
        },
        active: function (Q) {
          return Q === Q.ownerDocument.activeElement;
        },
        contains: function (S, R, Q) {
          return (S.textContent || S.innerText || a(S)).indexOf(Q[3]) >= 0;
        },
      },
      setFilters: {
        first: function (R, Q) {
          return Q === 0;
        },
        last: function (S, R, Q, U) {
          return R === U.length - 1;
        },
        even: function (R, Q) {
          return Q % 2 === 0;
        },
        odd: function (R, Q) {
          return Q % 2 === 1;
        },
        lt: function (S, R, Q) {
          return R < Q[3] - 0;
        },
        gt: function (S, R, Q) {
          return R > Q[3] - 0;
        },
        nth: function (S, R, Q) {
          return Q[3] - 0 === R;
        },
        eq: function (S, R, Q) {
          return Q[3] - 0 === R;
        },
      },
      filter: {
        PSEUDO: function (S, X, W, Z) {
          var Q = X[1],
            R = F.filters[Q];
          if (R) {
            return R(S, W, X, Z);
          } else {
            if (Q === "not") {
              var U = X[3],
                V = 0,
                Y = U.length;
              for (; V < Y; V++) {
                if (U[V] === S) {
                  return false;
                }
              }
              return true;
            } else {
              L.error(Q);
            }
          }
        },
        CHILD: function (S, V) {
          var U,
            ab,
            X,
            aa,
            Q,
            W,
            Z,
            Y = V[1],
            R = S;
          switch (Y) {
            case "only":
            case "first":
              while ((R = R.previousSibling)) {
                if (R.nodeType === 1) {
                  return false;
                }
              }
              if (Y === "first") {
                return true;
              }
              R = S;
            case "last":
              while ((R = R.nextSibling)) {
                if (R.nodeType === 1) {
                  return false;
                }
              }
              return true;
            case "nth":
              U = V[2];
              ab = V[3];
              if (U === 1 && ab === 0) {
                return true;
              }
              X = V[0];
              aa = S.parentNode;
              if (aa && (aa[O] !== X || !S.nodeIndex)) {
                W = 0;
                for (R = aa.firstChild; R; R = R.nextSibling) {
                  if (R.nodeType === 1) {
                    R.nodeIndex = ++W;
                  }
                }
                aa[O] = X;
              }
              Z = S.nodeIndex - ab;
              if (U === 0) {
                return Z === 0;
              } else {
                return Z % U === 0 && Z / U >= 0;
              }
          }
        },
        ID: i
          ? function (R, Q) {
              return R.nodeType === 1 && R.getAttribute("id") === Q;
            }
          : function (S, Q) {
              var R = typeof S.getAttributeNode !== A && S.getAttributeNode("id");
              return S.nodeType === 1 && R && R.nodeValue === Q;
            },
        TAG: function (R, Q) {
          return (Q === "*" && R.nodeType === 1) || (!!R.nodeName && R.nodeName.toLowerCase() === Q);
        },
        CLASS: function (R, Q) {
          return (" " + (R.className || R.getAttribute("class")) + " ").indexOf(Q) > -1;
        },
        ATTR: function (W, U) {
          var S = U[1],
            Q = L.attr ? L.attr(W, S) : F.attrHandle[S] ? F.attrHandle[S](W) : W[S] != null ? W[S] : W.getAttribute(S),
            X = Q + "",
            V = U[2],
            R = U[4];
          return Q == null
            ? V === "!="
            : !V && L.attr
              ? Q != null
              : V === "="
                ? X === R
                : V === "*="
                  ? X.indexOf(R) >= 0
                  : V === "~="
                    ? (" " + X + " ").indexOf(R) >= 0
                    : !R
                      ? X && Q !== false
                      : V === "!="
                        ? X !== R
                        : V === "^="
                          ? X.indexOf(R) === 0
                          : V === "$="
                            ? X.substr(X.length - R.length) === R
                            : V === "|="
                              ? X === R || X.substr(0, R.length + 1) === R + "-"
                              : false;
        },
        POS: function (V, R, S, W) {
          var Q = R[2],
            U = F.setFilters[Q];
          if (U) {
            return U(V, S, R, W);
          }
        },
      },
    });
    if (C) {
      F.order.splice(1, 0, "CLASS");
      F.find.CLASS = function (R, S, Q) {
        if (typeof S.getElementsByClassName !== A && !Q) {
          return S.getElementsByClassName(R[1]);
        }
      };
    }
    var p, b;
    if (j.compareDocumentPosition) {
      p = function (R, Q) {
        if (R === Q) {
          l = true;
          return 0;
        }
        if (!R.compareDocumentPosition || !Q.compareDocumentPosition) {
          return R.compareDocumentPosition ? -1 : 1;
        }
        return R.compareDocumentPosition(Q) & 4 ? -1 : 1;
      };
    } else {
      p = function (Z, Y) {
        if (Z === Y) {
          l = true;
          return 0;
        } else {
          if (Z.sourceIndex && Y.sourceIndex) {
            return Z.sourceIndex - Y.sourceIndex;
          }
        }
        var W,
          R,
          S = [],
          Q = [],
          V = Z.parentNode,
          X = Y.parentNode,
          aa = V;
        if (V === X) {
          return b(Z, Y);
        } else {
          if (!V) {
            return -1;
          } else {
            if (!X) {
              return 1;
            }
          }
        }
        while (aa) {
          S.unshift(aa);
          aa = aa.parentNode;
        }
        aa = X;
        while (aa) {
          Q.unshift(aa);
          aa = aa.parentNode;
        }
        W = S.length;
        R = Q.length;
        for (var U = 0; U < W && U < R; U++) {
          if (S[U] !== Q[U]) {
            return b(S[U], Q[U]);
          }
        }
        return U === W ? b(Z, Q[U], -1) : b(S[U], Y, 1);
      };
      b = function (R, Q, S) {
        if (R === Q) {
          return S;
        }
        var U = R.nextSibling;
        while (U) {
          if (U === Q) {
            return -1;
          }
          U = U.nextSibling;
        }
        return 1;
      };
    }
    if (e.querySelectorAll) {
      (function () {
        var V = N,
          U = "__sizzle__",
          S = /^\s*[+~]/,
          R = /'/g,
          Q = [];
        G(function (W) {
          W.innerHTML = "<select><option selected></option></select>";
          if (!W.querySelectorAll("[selected]").length) {
            Q.push("\\[[\\x20\\t\\n\\r\\f]*(?:checked|disabled|ismap|multiple|readonly|selected|value)");
          }
          if (!W.querySelectorAll(":checked").length) {
            Q.push(":checked");
          }
        });
        G(function (W) {
          W.innerHTML = "<p class=''></p>";
          if (W.querySelectorAll("[class^='']").length) {
            Q.push("[*^$]=[\\x20\\t\\n\\r\\f]*(?:\"\"|'')");
          }
          W.innerHTML = "<input type='hidden'>";
          if (!W.querySelectorAll(":enabled").length) {
            Q.push(":enabled", ":disabled");
          }
        });
        Q = Q.length && new RegExp(Q.join("|"));
        N = function (ac, X, ad, ae, ab) {
          if (!ae && !ab && (!Q || !Q.test(ac))) {
            if (X.nodeType === 9) {
              try {
                return t(X.querySelectorAll(ac), ad);
              } catch (aa) {}
            } else {
              if (X.nodeType === 1 && X.nodeName.toLowerCase() !== "object") {
                var Y = X,
                  Z = X.getAttribute("id"),
                  W = Z || U,
                  ag = X.parentNode,
                  af = S.test(ac);
                if (!Z) {
                  X.setAttribute("id", W);
                } else {
                  W = W.replace(R, "\\$&");
                }
                if (af && ag) {
                  X = ag;
                }
                try {
                  if (!af || ag) {
                    return t(X.querySelectorAll("[id='" + W + "'] " + ac), ad);
                  }
                } catch (aa) {
                } finally {
                  if (!Z) {
                    Y.removeAttribute("id");
                  }
                }
              }
            }
          }
          return V(ac, X, ad, ae, ab);
        };
      })();
    }
    function k(S, ab, Q, V) {
      var R,
        W,
        Z,
        aa,
        Y = o++,
        U = 0,
        X = ab.length;
      if (typeof Q === "string" && !P.test(Q)) {
        Q = Q.toLowerCase();
        aa = Q;
      }
      for (; U < X; U++) {
        R = ab[U];
        if (R) {
          W = false;
          R = R[S];
          while (R) {
            if (R[O] === Y) {
              W = ab[R.sizset];
              break;
            }
            Z = R.nodeType === 1;
            if (Z && !V) {
              R[O] = Y;
              R.sizset = U;
            }
            if (aa) {
              if (R.nodeName.toLowerCase() === Q) {
                W = R;
                break;
              }
            } else {
              if (Z) {
                if (typeof Q !== "string") {
                  if (R === Q) {
                    W = true;
                    break;
                  }
                } else {
                  if (L.filter(Q, [R]).length > 0) {
                    W = R;
                    break;
                  }
                }
              }
            }
            R = R[S];
          }
          ab[U] = W;
        }
      }
    }
    var x = function (S, Q, X, R) {
      var W,
        Z = [],
        V = "",
        aa = Q.nodeType ? [Q] : Q,
        U = 0,
        Y = aa.length;
      while ((W = E.PSEUDO.exec(S))) {
        V += W[0];
        S = S.replace(E.PSEUDO, "");
      }
      if (F.relative[S]) {
        S += "*";
      }
      for (; U < Y; U++) {
        N(S, aa[U], Z, X, R);
      }
      return L.filter(V, Z);
    };
    M.Sizzle = baidu.sizzle = L;
  })(window);
  baidu.string.extend({
    decodeHTML: function () {
      var a = this.replace(/&quot;/g, '"')
        .replace(/&lt;/g, "<")
        .replace(/&gt;/g, ">")
        .replace(/&amp;/g, "&");
      return a.replace(/&#([\d]+);/g, function (c, b) {
        return String.fromCharCode(parseInt(b, 10));
      });
    },
  });
  baidu.string.extend({
    encodeHTML: function () {
      return this.replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
    },
  });
  baidu.string.filterFormat = function (c, a) {
    var b = Array.prototype.slice.call(arguments, 1),
      d = Object.prototype.toString;
    if (b.length) {
      b = b.length == 1 ? (a !== null && /\[object Array\]|\[object Object\]/.test(d.call(a)) ? a : b) : b;
      return c.replace(/#\{(.+?)\}/g, function (f, j) {
        var l, h, g, e, k;
        if (!b) {
          return "";
        }
        l = j.split("|");
        h = b[l[0]];
        if ("[object Function]" == d.call(h)) {
          h = h(l[0]);
        }
        for (g = 1, e = l.length; g < e; ++g) {
          k = baidu.string.filterFormat[l[g]];
          if ("[object Function]" == d.call(k)) {
            h = k(h);
          }
        }
        return "undefined" == typeof h || h === null ? "" : h;
      });
    }
    return c;
  };
  baidu.string.filterFormat.escapeJs = function (e) {
    if (!e || "string" != typeof e) {
      return e;
    }
    var d,
      a,
      b,
      c = [];
    for (d = 0, a = e.length; d < a; ++d) {
      b = e.charCodeAt(d);
      if (b > 255) {
        c.push(e.charAt(d));
      } else {
        c.push("\\x" + b.toString(16));
      }
    }
    return c.join("");
  };
  baidu.string.filterFormat.js = baidu.string.filterFormat.escapeJs;
  baidu.string.filterFormat.escapeString = function (a) {
    if (!a || "string" != typeof a) {
      return a;
    }
    return a.replace(/["'<>\\\/`]/g, function (b) {
      return "&#" + b.charCodeAt(0) + ";";
    });
  };
  baidu.string.filterFormat.e = baidu.string.filterFormat.escapeString;
  baidu.string.filterFormat.toInt = function (a) {
    return parseInt(a, 10) || 0;
  };
  baidu.string.filterFormat.i = baidu.string.filterFormat.toInt;
  baidu.string.extend({
    format: function (a) {
      var c = this.valueOf(),
        b = Array.prototype.slice.call(arguments, 0),
        d = Object.prototype.toString;
      if (b.length) {
        b = b.length == 1 ? (a !== null && /\[object Array\]|\[object Object\]/.test(d.call(a)) ? a : b) : b;
        return c.replace(/#\{(.+?)\}/g, function (e, g) {
          var f = b[g];
          if ("[object Function]" == d.call(f)) {
            f = f(g);
          }
          return "undefined" == typeof f ? "" : f;
        });
      }
      return c;
    },
  });
  baidu.string.extend({
    getByteLength: function () {
      return this.replace(/[^\x00-\xff]/g, "ci").length;
    },
  });
  baidu.string.extend({
    stripTags: function () {
      return (this || "").replace(/<[^>]+>/g, "");
    },
  });
  baidu.string.extend({
    subByte: function (a, b) {
      baidu.check("^(?:number(?:,(?:string|number))?)$", "baidu.string.subByte");
      var c = this.valueOf();
      b = b || "";
      if (a < 0 || baidu.string(c).getByteLength() <= a) {
        return c + b;
      }
      c = c
        .substr(0, a)
        .replace(/([^\x00-\xff])/g, "\x241 ")
        .substr(0, a)
        .replace(/[^\x00-\xff]$/, "")
        .replace(/([^\x00-\xff]) /g, "\x241");
      return c + b;
    },
  });
  baidu.string.extend({
    toHalfWidth: function () {
      return this.replace(/[\uFF01-\uFF5E]/g, function (a) {
        return String.fromCharCode(a.charCodeAt(0) - 65248);
      }).replace(/\u3000/g, " ");
    },
  });
  baidu.string.extend({
    wbr: function () {
      return this.replace(/(?:<[^>]+>)|(?:&#?[0-9a-z]{2,6};)|(.{1})/gi, "$&<wbr>").replace(/><wbr>/g, ">");
    },
  });
  baidu.swf = baidu.swf || {};
  baidu.swf.version = (function () {
    var h = navigator;
    if (h.plugins && h.mimeTypes.length) {
      var d = h.plugins["Shockwave Flash"];
      if (d && d.description) {
        return d.description.replace(/([a-zA-Z]|\s)+/, "").replace(/(\s)+r/, ".") + ".0";
      }
    } else {
      if (window.ActiveXObject && !window.opera) {
        for (var b = 12; b >= 2; b--) {
          try {
            var g = new ActiveXObject("ShockwaveFlash.ShockwaveFlash." + b);
            if (g) {
              var a = g.GetVariable("$version");
              return a.replace(/WIN/g, "").replace(/,/g, ".");
            }
          } catch (f) {}
        }
      }
    }
  })();
  baidu.swf.createHTML = function (s) {
    s = s || {};
    var j = baidu.swf.version,
      g = s["ver"] || "6.0.0",
      f,
      d,
      e,
      c,
      h,
      r,
      a = {},
      o = baidu.string.encodeHTML;
    for (c in s) {
      a[c] = s[c];
    }
    s = a;
    if (j) {
      j = j.split(".");
      g = g.split(".");
      for (e = 0; e < 3; e++) {
        f = parseInt(j[e], 10);
        d = parseInt(g[e], 10);
        if (d < f) {
          break;
        } else {
          if (d > f) {
            return "";
          }
        }
      }
    } else {
      return "";
    }
    var m = s["vars"],
      l = ["classid", "codebase", "id", "width", "height", "align"];
    s["align"] = s["align"] || "middle";
    s["classid"] = "clsid:d27cdb6e-ae6d-11cf-96b8-444553540000";
    s["codebase"] = "http://fpdownload.macromedia.com/pub/shockwave/cabs/flash/swflash.cab#version=6,0,0,0";
    s["movie"] = s["url"] || "";
    delete s["vars"];
    delete s["url"];
    if ("string" == typeof m) {
      s["flashvars"] = m;
    } else {
      var p = [];
      for (c in m) {
        r = m[c];
        p.push(c + "=" + encodeURIComponent(r));
      }
      s["flashvars"] = p.join("&");
    }
    var n = ["<object "];
    for (e = 0, h = l.length; e < h; e++) {
      r = l[e];
      n.push(" ", r, '="', o(s[r]), '"');
    }
    n.push(">");
    var b = {
      wmode: 1,
      scale: 1,
      quality: 1,
      play: 1,
      loop: 1,
      menu: 1,
      salign: 1,
      bgcolor: 1,
      base: 1,
      allowscriptaccess: 1,
      allownetworking: 1,
      allowfullscreen: 1,
      seamlesstabbing: 1,
      devicefont: 1,
      swliveconnect: 1,
      flashvars: 1,
      movie: 1,
    };
    for (c in s) {
      r = s[c];
      c = c.toLowerCase();
      if (b[c] && (r || r === false || r === 0)) {
        n.push('<param name="' + c + '" value="' + o(r) + '" />');
      }
    }
    s["src"] = s["movie"];
    s["name"] = s["id"];
    delete s["id"];
    delete s["movie"];
    delete s["classid"];
    delete s["codebase"];
    s["type"] = "application/x-shockwave-flash";
    s["pluginspage"] = "http://www.macromedia.com/go/getflashplayer";
    n.push("<embed");
    var q;
    for (c in s) {
      r = s[c];
      if (r || r === false || r === 0) {
        if (new RegExp("^salign\x24", "i").test(c)) {
          q = r;
          continue;
        }
        n.push(" ", c, '="', o(r), '"');
      }
    }
    if (q) {
      n.push(' salign="', o(q), '"');
    }
    n.push("></embed></object>");
    return n.join("");
  };
  baidu.swf.create = function (a, c) {
    a = a || {};
    var b = baidu.swf.createHTML(a) || a["errorMessage"] || "";
    if (c && "string" == typeof c) {
      c = document.getElementById(c);
    }
    baidu.dom.insertHTML(c || document.body, "beforeEnd", b);
  };
  baidu.swf.getMovie = function (c) {
    var a = document[c],
      b;
    return baidu.browser.ie == 9
      ? a && a.length
        ? (b = baidu.array.remove(baidu.lang.toArray(a), function (d) {
            return d.tagName.toLowerCase() != "embed";
          })).length == 1
          ? b[0]
          : b
        : a
      : a || window[c];
  };
  baidu.swf.Proxy = function (f, c, d) {
    var b = this,
      a = (this._flash = baidu.swf.getMovie(f)),
      e;
    if (!c) {
      return this;
    }
    e = setInterval(function () {
      try {
        if (a[c]) {
          b._initialized = true;
          clearInterval(e);
          if (d) {
            d();
          }
        }
      } catch (g) {}
    }, 100);
  };
  baidu.swf.Proxy.prototype.getFlash = function () {
    return this._flash;
  };
  baidu.swf.Proxy.prototype.isReady = function () {
    return !!this._initialized;
  };
  baidu.swf.Proxy.prototype.call = function (a, f) {
    try {
      var c = this.getFlash(),
        b = Array.prototype.slice.call(arguments);
      b.shift();
      if (c[a]) {
        c[a].apply(c, b);
      }
    } catch (d) {}
  };
  baidu.url.getQueryValue = function (b, c) {
    var d = new RegExp("(^|&|\\?|#)" + baidu.string.escapeReg(c) + "=([^&#]*)(&|\x24|#)", "");
    var a = b.match(d);
    if (a) {
      return a[2];
    }
    return null;
  };
  baidu.url.jsonToQuery = function (c, e) {
    var a = [],
      d,
      b =
        e ||
        function (f) {
          return baidu.url.escapeSymbol(f);
        };
    baidu.object.each(c, function (g, f) {
      if (baidu.lang.isArray(g)) {
        d = g.length;
        while (d--) {
          a.push(f + "=" + b(g[d], f));
        }
      } else {
        a.push(f + "=" + b(g, f));
      }
    });
    return a.join("&");
  };
  baidu.url.queryToJson = function (c) {
    var h = c.substr(c.lastIndexOf("?") + 1).split("&"),
      a = h.length,
      b = null,
      f,
      e,
      g;
    for (var d = 0; d < a; d++) {
      f = h[d].split("=");
      if (f.length < 2) {
        continue;
      }
      !b && (b = {});
      e = f[0];
      g = f[1];
      f = b[e];
      if (!f) {
        b[e] = g;
      } else {
        if (baidu.lang.isArray(f)) {
          f.push(g);
        } else {
          b[e] = [f, g];
        }
      }
    }
    return b;
  };
  (function () {
    var f = /^(?:\{.*\}|\[.*\])$/,
      a = /([A-Z])/g,
      c = /^-ms-/,
      e = /-([\da-z])/gi,
      b = function (h, i) {
        return (i + "").toUpperCase();
      };
    baidu.extend(baidu._util_, {
      camelCase: function (h) {
        return h.replace(c, "ms-").replace(e, b);
      },
    });
    baidu.extend(baidu._util_, {
      cache: {},
      deletedIds: [],
      uuid: 0,
      expando: "baidu" + ("2.0.0" + Math.random()).replace(/\D/g, ""),
      noData: { embed: true, object: "clsid:D27CDB6E-AE6D-11cf-96B8-444553540000", applet: true },
      hasData: function (h) {
        h = h.nodeType ? baidu._util_.cache[h[baidu._util_.expando]] : h[baidu._util_.expando];
        return !!h && !d(h);
      },
      data: function (k, i, m, l) {
        if (!baidu._util_.acceptData(k)) {
          return;
        }
        var n,
          p,
          q = baidu._util_.expando,
          o = typeof i === "string",
          r = k.nodeType,
          h = r ? baidu._util_.cache : k,
          j = r ? k[q] : k[q] && q;
        if ((!j || !h[j] || (!l && !h[j].data)) && o && m === undefined) {
          return;
        }
        if (!j) {
          if (r) {
            k[q] = j = baidu._util_.deletedIds.pop() || ++baidu._util_.uuid;
          } else {
            j = q;
          }
        }
        if (!h[j]) {
          h[j] = {};
          if (!r) {
            h[j].toJSON = function () {};
          }
        }
        if (typeof i === "object" || typeof i === "function") {
          if (l) {
            h[j] = baidu.extend(h[j], i);
          } else {
            h[j].data = baidu.extend(h[j].data, i);
          }
        }
        n = h[j];
        if (!l) {
          if (!n.data) {
            n.data = {};
          }
          n = n.data;
        }
        if (m !== undefined) {
          n[baidu._util_.camelCase(i)] = m;
        }
        if (o) {
          p = n[i];
          if (p == null) {
            p = n[baidu._util_.camelCase(i)];
          }
        } else {
          p = n;
        }
        return p;
      },
      removeData: function (m, j, n) {
        if (!baidu._util_.acceptData(m)) {
          return;
        }
        var q,
          p,
          o,
          r = m.nodeType,
          h = r ? baidu._util_.cache : m,
          k = r ? m[baidu._util_.expando] : baidu._util_.expando;
        if (!h[k]) {
          return;
        }
        if (j) {
          q = n ? h[k] : h[k].data;
          if (q) {
            if (!baidu.isArray(j)) {
              if (j in q) {
                j = [j];
              } else {
                j = baidu._util_.camelCase(j);
                if (j in q) {
                  j = [j];
                } else {
                  j = j.split(" ");
                }
              }
            }
            for (p = 0, o = j.length; p < o; p++) {
              delete q[j[p]];
            }
            if (!(n ? d : baidu.object.isEmpty)(q)) {
              return;
            }
          }
        }
        if (!n) {
          delete h[k].data;
          if (!d(h[k])) {
            return;
          }
        }
        if (r) {
          baidu._util_.cleanData([m], true);
        } else {
          if (baidu.support.deleteExpando || h != h.window) {
            delete h[k];
          } else {
            h[k] = null;
          }
        }
      },
      _data: function (i, h, j) {
        return baidu._util_.data(i, h, j, true);
      },
      acceptData: function (i) {
        var h = i.nodeName && baidu._util_.noData[i.nodeName.toLowerCase()];
        return !h || (h !== true && i.getAttribute("classid") === h);
      },
    });
    function g(j, i, k) {
      if (k === undefined && j.nodeType === 1) {
        var h = "data-" + i.replace(a, "-$1").toLowerCase();
        k = j.getAttribute(h);
        if (typeof k === "string") {
          try {
            k =
              k === "true"
                ? true
                : k === "false"
                  ? false
                  : k === "null"
                    ? null
                    : baidu.isNumber(k)
                      ? +k
                      : f.test(k)
                        ? baidu.json.parse(k)
                        : k;
          } catch (l) {}
          baidu._util_.data(j, i, k);
        } else {
          k = undefined;
        }
      }
      return k;
    }
    function d(i) {
      var h;
      for (h in i) {
        if (h === "data" && baidu.object.isEmpty(i[h])) {
          continue;
        }
        if (h !== "toJSON") {
          return false;
        }
      }
      return true;
    }
  })();
  baidu.array = baidu.array || {};
  baidu.dom = baidu.dom || {};
  baidu.addClass = baidu.dom.addClass || {};
  baidu.g = baidu.G = baidu.dom.g || {};
  baidu.getAttr = baidu.dom.getAttr || {};
  baidu.getStyle = baidu.dom.getStyle || {};
  baidu.hide = baidu.dom.hide || {};
  baidu.insertHTML = baidu.dom.insertHTML || {};
  baidu.q = baidu.Q = baidu.dom.q || {};
  baidu.removeClass = baidu.dom.removeClass || {};
  baidu.setAttr = baidu.dom.setAttr || {};
  baidu.setAttrs = baidu.dom.setAttrs || {};
  baidu.dom.setOuterHeight = baidu.dom.setBorderBoxHeight || {};
  baidu.dom.setOuterWidth = baidu.dom.setBorderBoxWidth || {};
  baidu.setStyle = baidu.dom.setStyle || {};
  baidu.setStyles = baidu.dom.setStyles || {};
  baidu.show = baidu.dom.show || {};
  baidu.e = baidu.element = baidu.element || {};
  baidu.event = baidu.event || {};
  baidu.on = baidu.event.on || {};
  baidu.un = baidu.event.un || {};
  baidu.lang = baidu.lang || {};
  baidu.inherits = baidu.lang.inherits || {};
  baidu.object = baidu.object || {};
  baidu.string = baidu.string || {};
  baidu.decodeHTML = baidu.string.decodeHTML || {};
  baidu.encodeHTML = baidu.string.encodeHTML || {};
  baidu.format = baidu.string.format || {};
  baidu.trim = baidu.string.trim || {};
  var passport = passport || window.passport || {};
  !(function (a) {
    var t = "https://ufosdk.baidu.com/bailing/getEntryPath/eG8Bo9LKhwNaa+UzF4RrYeKgloEGHe8h40KVmAE9l3E=",
      o = "https://ufosdk.baidu.com/bailingPC/getEntryPath/eG8Bo9LKhwNaa+UzF4RrYU-3KF8SAV35kX+Zqjzyp8U=";
    ((a.MobileDukeUrl = t), (a.PcDukeUrl = o));
  })(passport);
  var passport = passport || window.passport || {},
    baidu = passport.tangramInst || baidu || window.baidu;
  !(function (t) {
    function p(t) {
      var p = a[t];
      if (!p) return "";
      if ("string" == typeof p) return p;
      try {
        var s = location.hostname;
        for (var o in p) if (s === o || s.endsWith("." + o)) return p[o];
      } catch (r) {}
      return "";
    }
    var s,
      o = document.location.protocol.toLowerCase();
    ((s =
      passport && passport.ieVersion && passport.ieVersion() <= 8
        ? "https:" === o
          ? "https://passport.baidu.com"
          : "http://passport.baidu.com"
        : "https:" === o
          ? "https://ppui-static-pc.cdn.bcebos.com"
          : "http://ppui-static-pc.cdn.bcebos.com"),
      (t.apiDomain = { staticDomain: s }));
    var a = {
      yiyan_intl: "https://passport.ideachat.com",
      mcpworld: "https://passport.mcpworld.com",
      ares_efficiency: "https://passport.mcpworld.com",
      callcrab: "https://passport.callcrab.com",
      genflowtwo: {
        "kukukuku.cn": "https://passport.kukukuku.cn",
        "kukuai.cn": "https://passport.kukuai.cn",
        "genflow.cn": "https://passport.genflow.cn",
      },
    };
    t.requestDomain || (t.requestDomain = p(passport.tpl) || "https://passport.baidu.com");
  })(passport);
  var magic = null;
  if ("function" != typeof magic) var magic = function () {};
  magic._baiduInstName = magic._baiduInstName || "bdInst_" + new Date().getTime();
  var baiduInstance = baiduInstance || baidu.baiduInstance || window.baiduInstance;
  ((window[magic._baiduInstName] = window[magic._baiduInstName] || baiduInstance),
    (magic.resourcePath = ""),
    (magic.skinName = "default"),
    (magic.version = "1.0.0.0"),
    /msie 6/i.test(navigator.userAgent) && document.execCommand("BackgroundImageCache", !1, !0),
    (baidu.form = baidu.form || {}),
    (baidu.url = baidu.url || {}),
    (baidu.url.escapeSymbol =
      baidu.url.escapeSymbol ||
      function (e) {
        return String(e).replace(/[#%&+=\/\\\ \　\f\r\n\t]/g, function (e) {
          return "%" + (256 + e.charCodeAt()).toString(16).substring(1).toUpperCase();
        });
      }),
    (baidu.form.json =
      baidu.form.json ||
      function (e, i) {
        function t(e, i) {
          var t = g[e];
          t ? (t.push || (g[e] = [t]), g[e].push(i)) : (g[e] = i);
        }
        for (
          var a,
            n,
            s,
            r,
            u,
            c,
            o,
            d,
            l = e.elements,
            i =
              i ||
              function (e) {
                return e;
              },
            g = {},
            m = 0,
            b = l.length;
          b > m;
          m++
        )
          if (((a = l[m]), (s = a.name), !a.disabled && s))
            switch (((n = a.type), (r = baidu.url.escapeSymbol(a.value)), n)) {
              case "radio":
              case "checkbox":
                a.checked && t(s, i(r, s));
                break;
              case "textarea":
              case "text":
              case "password":
              case "hidden":
              case "file":
              case "select-one":
                t(s, i(r, s));
                break;
              case "select-multiple":
                for (u = a.options, o = u.length, c = 0; o > c; c++) ((d = u[c]), d.selected && t(s, i(d.value, s)));
            }
        return g;
      }),
    (magic.Base = function () {
      (baidu.lang.Class.call(this), (this._ids = {}), (this._eid = this.guid + "__"));
    }),
    baidu.lang.inherits(magic.Base, baidu.lang.Class, "magic.Base").extend({
      getElement: function (e) {
        return document.getElementById(this.$getId(e));
      },
      getElements: function () {
        var e = {},
          i = this._ids;
        for (var t in i) e[t] = this.getElement(t);
        return e;
      },
      $getId: function (e) {
        return ((e = baidu.lang.isString(e) ? e : ""), this._ids[e] || (this._ids[e] = this._eid + e));
      },
      $mappingDom: function (e, i) {
        return (
          baidu.lang.isString(i)
            ? (this._ids[e] = i)
            : i && i.nodeType && (i.id ? (this._ids[e] = i.id) : (i.id = this.$getId(e))),
          this
        );
      },
      $hide: function (e) {
        return (
          ("string" == (typeof e).toLowerCase() || "" === e) && (e = this.getElement(e)),
          e && e.style && ((e.style.display = "none"), (e.style.visibility = "hidden")),
          this
        );
      },
      $show: function (e) {
        return (
          ("string" == (typeof e).toLowerCase() || "" === e) && (e = this.getElement(e)),
          e && e.style && ((e.style.display = "block"), (e.style.visibility = "visible"), (e.style.opacity = "1")),
          this
        );
      },
      $dispose: function () {
        this.fire("ondispose") && baidu.lang.Class.prototype.dispose.call(this);
      },
    }),
    (magic.control = magic.control || {}),
    (function () {
      function e(e, i) {
        var a,
          n = e.getAttribute(i),
          s = !1;
        if (n && (a = n.match(t[0]))) {
          s = {};
          for (var r, u = 0; u < a.length; u++)
            ((r = a[u].match(t[1])),
              !isNaN(r[2]) && (r[2] = +r[2]),
              t[2].test(r[2]) && (r[2] = r[2].replace(t[3], "$2")),
              t[4].test(r[2]) && (r[2] = t[5].test(r[2])),
              (s[r[1]] = r[2]));
        }
        return s;
      }
      function i(i, t) {
        var a = e(i, "tang-event");
        if (a)
          for (var n in a) {
            var s = a[n].substr(1);
            (s.indexOf("(") < 0 && (s += "()"),
              baidu
                .dom(i)
                .on(
                  n,
                  new Function(magic._baiduInstName + "('" + t + "') && " + magic._baiduInstName + "('" + t + "')" + s),
                ));
          }
      }
      magic.setup =
        magic.setup ||
        function (t, a, n) {
          var s = e(t, "tang-param") || {};
          for (var r in n) s[r] = n[r];
          var u = new a(s);
          (u.$mappingDom("", t), i(t, u.guid));
          for (var c = t.getElementsByTagName("*"), r = c.length - 1; r >= 0; r--) i(c[r], u.guid);
          return u;
        };
      var t = [
        /\b[\w\$\-]+\s*:\s*[^;]+/g,
        /([\w\$\-]+)\s*:\s*([^;]+)\s*/,
        /\'|\"/,
        /^\s*(\'|\")([^\1]*)\1\s*/,
        /^(true|false)\s*$/i,
        /\btrue\b/i,
      ];
    })());
  ((passport = passport || {}),
    (passport.lib = passport.lib || {}),
    (passport.lib.RSAExport = {}),
    (function (t) {
      function e(t, e, i) {
        null != t &&
          ("number" == typeof t
            ? this.fromNumber(t, e, i)
            : null == e && "string" != typeof t
              ? this.fromString(t, 256)
              : this.fromString(t, e));
      }
      function i() {
        return new e(null);
      }
      function r(t, e, i, r, s, n) {
        for (; --n >= 0;) {
          var o = e * this[t++] + i[r] + s;
          ((s = Math.floor(o / 67108864)), (i[r++] = 67108863 & o));
        }
        return s;
      }
      function s(t, e, i, r, s, n) {
        for (var o = 32767 & e, h = e >> 15; --n >= 0;) {
          var a = 32767 & this[t],
            u = this[t++] >> 15,
            c = h * a + u * o;
          ((a = o * a + ((32767 & c) << 15) + i[r] + (1073741823 & s)),
            (s = (a >>> 30) + (c >>> 15) + h * u + (s >>> 30)),
            (i[r++] = 1073741823 & a));
        }
        return s;
      }
      function n(t, e, i, r, s, n) {
        for (var o = 16383 & e, h = e >> 14; --n >= 0;) {
          var a = 16383 & this[t],
            u = this[t++] >> 14,
            c = h * a + u * o;
          ((a = o * a + ((16383 & c) << 14) + i[r] + s), (s = (a >> 28) + (c >> 14) + h * u), (i[r++] = 268435455 & a));
        }
        return s;
      }
      function o(t) {
        return Ki.charAt(t);
      }
      function h(t, e) {
        var i = Ui[t.charCodeAt(e)];
        return null == i ? -1 : i;
      }
      function a(t) {
        for (var e = this.t - 1; e >= 0; --e) t[e] = this[e];
        ((t.t = this.t), (t.s = this.s));
      }
      function u(t) {
        ((this.t = 1), (this.s = 0 > t ? -1 : 0), t > 0 ? (this[0] = t) : -1 > t ? (this[0] = t + DV) : (this.t = 0));
      }
      function c(t) {
        var e = i();
        return (e.fromInt(t), e);
      }
      function f(t, i) {
        var r;
        if (16 == i) r = 4;
        else if (8 == i) r = 3;
        else if (256 == i) r = 8;
        else if (2 == i) r = 1;
        else if (32 == i) r = 5;
        else {
          if (4 != i) return void this.fromRadix(t, i);
          r = 2;
        }
        ((this.t = 0), (this.s = 0));
        for (var s = t.length, n = !1, o = 0; --s >= 0;) {
          var a = 8 == r ? 255 & t[s] : h(t, s);
          0 > a
            ? "-" == t.charAt(s) && (n = !0)
            : ((n = !1),
              0 == o
                ? (this[this.t++] = a)
                : o + r > this.DB
                  ? ((this[this.t - 1] |= (a & ((1 << (this.DB - o)) - 1)) << o), (this[this.t++] = a >> (this.DB - o)))
                  : (this[this.t - 1] |= a << o),
              (o += r),
              o >= this.DB && (o -= this.DB));
        }
        (8 == r && 0 != (128 & t[0]) && ((this.s = -1), o > 0 && (this[this.t - 1] |= ((1 << (this.DB - o)) - 1) << o)),
          this.clamp(),
          n && e.ZERO.subTo(this, this));
      }
      function p() {
        for (var t = this.s & this.DM; this.t > 0 && this[this.t - 1] == t;) --this.t;
      }
      function l(t) {
        if (this.s < 0) return "-" + this.negate().toString(t);
        var e;
        if (16 == t) e = 4;
        else if (8 == t) e = 3;
        else if (2 == t) e = 1;
        else if (32 == t) e = 5;
        else {
          if (4 != t) return this.toRadix(t);
          e = 2;
        }
        var i,
          r = (1 << e) - 1,
          s = !1,
          n = "",
          h = this.t,
          a = this.DB - ((h * this.DB) % e);
        if (h-- > 0)
          for (a < this.DB && (i = this[h] >> a) > 0 && ((s = !0), (n = o(i))); h >= 0;)
            (e > a
              ? ((i = (this[h] & ((1 << a) - 1)) << (e - a)), (i |= this[--h] >> (a += this.DB - e)))
              : ((i = (this[h] >> (a -= e)) & r), 0 >= a && ((a += this.DB), --h)),
              i > 0 && (s = !0),
              s && (n += o(i)));
        return s ? n : "0";
      }
      function d() {
        var t = i();
        return (e.ZERO.subTo(this, t), t);
      }
      function g() {
        return this.s < 0 ? this.negate() : this;
      }
      function m(t) {
        var e = this.s - t.s;
        if (0 != e) return e;
        var i = this.t;
        if (((e = i - t.t), 0 != e)) return this.s < 0 ? -e : e;
        for (; --i >= 0;) if (0 != (e = this[i] - t[i])) return e;
        return 0;
      }
      function y(t) {
        var e,
          i = 1;
        return (
          0 != (e = t >>> 16) && ((t = e), (i += 16)),
          0 != (e = t >> 8) && ((t = e), (i += 8)),
          0 != (e = t >> 4) && ((t = e), (i += 4)),
          0 != (e = t >> 2) && ((t = e), (i += 2)),
          0 != (e = t >> 1) && ((t = e), (i += 1)),
          i
        );
      }
      function b() {
        return this.t <= 0 ? 0 : this.DB * (this.t - 1) + y(this[this.t - 1] ^ (this.s & this.DM));
      }
      function T(t, e) {
        var i;
        for (i = this.t - 1; i >= 0; --i) e[i + t] = this[i];
        for (i = t - 1; i >= 0; --i) e[i] = 0;
        ((e.t = this.t + t), (e.s = this.s));
      }
      function S(t, e) {
        for (var i = t; i < this.t; ++i) e[i - t] = this[i];
        ((e.t = Math.max(this.t - t, 0)), (e.s = this.s));
      }
      function R(t, e) {
        var i,
          r = t % this.DB,
          s = this.DB - r,
          n = (1 << s) - 1,
          o = Math.floor(t / this.DB),
          h = (this.s << r) & this.DM;
        for (i = this.t - 1; i >= 0; --i) ((e[i + o + 1] = (this[i] >> s) | h), (h = (this[i] & n) << r));
        for (i = o - 1; i >= 0; --i) e[i] = 0;
        ((e[o] = h), (e.t = this.t + o + 1), (e.s = this.s), e.clamp());
      }
      function E(t, e) {
        e.s = this.s;
        var i = Math.floor(t / this.DB);
        if (i >= this.t) return void (e.t = 0);
        var r = t % this.DB,
          s = this.DB - r,
          n = (1 << r) - 1;
        e[0] = this[i] >> r;
        for (var o = i + 1; o < this.t; ++o) ((e[o - i - 1] |= (this[o] & n) << s), (e[o - i] = this[o] >> r));
        (r > 0 && (e[this.t - i - 1] |= (this.s & n) << s), (e.t = this.t - i), e.clamp());
      }
      function D(t, e) {
        for (var i = 0, r = 0, s = Math.min(t.t, this.t); s > i;)
          ((r += this[i] - t[i]), (e[i++] = r & this.DM), (r >>= this.DB));
        if (t.t < this.t) {
          for (r -= t.s; i < this.t;) ((r += this[i]), (e[i++] = r & this.DM), (r >>= this.DB));
          r += this.s;
        } else {
          for (r += this.s; i < t.t;) ((r -= t[i]), (e[i++] = r & this.DM), (r >>= this.DB));
          r -= t.s;
        }
        ((e.s = 0 > r ? -1 : 0), -1 > r ? (e[i++] = this.DV + r) : r > 0 && (e[i++] = r), (e.t = i), e.clamp());
      }
      function w(t, i) {
        var r = this.abs(),
          s = t.abs(),
          n = r.t;
        for (i.t = n + s.t; --n >= 0;) i[n] = 0;
        for (n = 0; n < s.t; ++n) i[n + r.t] = r.am(0, s[n], i, n, 0, r.t);
        ((i.s = 0), i.clamp(), this.s != t.s && e.ZERO.subTo(i, i));
      }
      function x(t) {
        for (var e = this.abs(), i = (t.t = 2 * e.t); --i >= 0;) t[i] = 0;
        for (i = 0; i < e.t - 1; ++i) {
          var r = e.am(i, e[i], t, 2 * i, 0, 1);
          (t[i + e.t] += e.am(i + 1, 2 * e[i], t, 2 * i + 1, r, e.t - i - 1)) >= e.DV &&
            ((t[i + e.t] -= e.DV), (t[i + e.t + 1] = 1));
        }
        (t.t > 0 && (t[t.t - 1] += e.am(i, e[i], t, 2 * i, 0, 1)), (t.s = 0), t.clamp());
      }
      function A(t, r, s) {
        var n = t.abs();
        if (!(n.t <= 0)) {
          var o = this.abs();
          if (o.t < n.t) return (null != r && r.fromInt(0), void (null != s && this.copyTo(s)));
          null == s && (s = i());
          var h = i(),
            a = this.s,
            u = t.s,
            c = this.DB - y(n[n.t - 1]);
          c > 0 ? (n.lShiftTo(c, h), o.lShiftTo(c, s)) : (n.copyTo(h), o.copyTo(s));
          var f = h.t,
            p = h[f - 1];
          if (0 != p) {
            var l = p * (1 << this.F1) + (f > 1 ? h[f - 2] >> this.F2 : 0),
              d = this.FV / l,
              g = (1 << this.F1) / l,
              m = 1 << this.F2,
              v = s.t,
              b = v - f,
              T = null == r ? i() : r;
            for (
              h.dlShiftTo(b, T),
                s.compareTo(T) >= 0 && ((s[s.t++] = 1), s.subTo(T, s)),
                e.ONE.dlShiftTo(f, T),
                T.subTo(h, h);
              h.t < f;
            )
              h[h.t++] = 0;
            for (; --b >= 0;) {
              var S = s[--v] == p ? this.DM : Math.floor(s[v] * d + (s[v - 1] + m) * g);
              if ((s[v] += h.am(0, S, s, b, 0, f)) < S)
                for (h.dlShiftTo(b, T), s.subTo(T, s); s[v] < --S;) s.subTo(T, s);
            }
            (null != r && (s.drShiftTo(f, r), a != u && e.ZERO.subTo(r, r)),
              (s.t = f),
              s.clamp(),
              c > 0 && s.rShiftTo(c, s),
              0 > a && e.ZERO.subTo(s, s));
          }
        }
      }
      function B(t) {
        var r = i();
        return (this.abs().divRemTo(t, null, r), this.s < 0 && r.compareTo(e.ZERO) > 0 && t.subTo(r, r), r);
      }
      function K(t) {
        this.m = t;
      }
      function U(t) {
        return t.s < 0 || t.compareTo(this.m) >= 0 ? t.mod(this.m) : t;
      }
      function O(t) {
        return t;
      }
      function V(t) {
        t.divRemTo(this.m, null, t);
      }
      function N(t, e, i) {
        (t.multiplyTo(e, i), this.reduce(i));
      }
      function J(t, e) {
        (t.squareTo(e), this.reduce(e));
      }
      function I() {
        if (this.t < 1) return 0;
        var t = this[0];
        if (0 == (1 & t)) return 0;
        var e = 3 & t;
        return (
          (e = (e * (2 - (15 & t) * e)) & 15),
          (e = (e * (2 - (255 & t) * e)) & 255),
          (e = (e * (2 - (((65535 & t) * e) & 65535))) & 65535),
          (e = (e * (2 - ((t * e) % this.DV))) % this.DV),
          e > 0 ? this.DV - e : -e
        );
      }
      function P(t) {
        ((this.m = t),
          (this.mp = t.invDigit()),
          (this.mpl = 32767 & this.mp),
          (this.mph = this.mp >> 15),
          (this.um = (1 << (t.DB - 15)) - 1),
          (this.mt2 = 2 * t.t));
      }
      function M(t) {
        var r = i();
        return (
          t.abs().dlShiftTo(this.m.t, r),
          r.divRemTo(this.m, null, r),
          t.s < 0 && r.compareTo(e.ZERO) > 0 && this.m.subTo(r, r),
          r
        );
      }
      function L(t) {
        var e = i();
        return (t.copyTo(e), this.reduce(e), e);
      }
      function q(t) {
        for (; t.t <= this.mt2;) t[t.t++] = 0;
        for (var e = 0; e < this.m.t; ++e) {
          var i = 32767 & t[e],
            r = (i * this.mpl + (((i * this.mph + (t[e] >> 15) * this.mpl) & this.um) << 15)) & t.DM;
          for (i = e + this.m.t, t[i] += this.m.am(0, r, t, e, 0, this.m.t); t[i] >= t.DV;) ((t[i] -= t.DV), t[++i]++);
        }
        (t.clamp(), t.drShiftTo(this.m.t, t), t.compareTo(this.m) >= 0 && t.subTo(this.m, t));
      }
      function H(t, e) {
        (t.squareTo(e), this.reduce(e));
      }
      function C(t, e, i) {
        (t.multiplyTo(e, i), this.reduce(i));
      }
      function j() {
        return 0 == (this.t > 0 ? 1 & this[0] : this.s);
      }
      function k(t, r) {
        if (t > 4294967295 || 1 > t) return e.ONE;
        var s = i(),
          n = i(),
          o = r.convert(this),
          h = y(t) - 1;
        for (o.copyTo(s); --h >= 0;)
          if ((r.sqrTo(s, n), (t & (1 << h)) > 0)) r.mulTo(n, o, s);
          else {
            var a = s;
            ((s = n), (n = a));
          }
        return r.revert(s);
      }
      function F(t, e) {
        var i;
        return ((i = 256 > t || e.isEven() ? new K(e) : new P(e)), this.exp(t, i));
      }
      function _() {
        var t = i();
        return (this.copyTo(t), t);
      }
      function z() {
        if (this.s < 0) {
          if (1 == this.t) return this[0] - this.DV;
          if (0 == this.t) return -1;
        } else {
          if (1 == this.t) return this[0];
          if (0 == this.t) return 0;
        }
        return ((this[1] & ((1 << (32 - this.DB)) - 1)) << this.DB) | this[0];
      }
      function Z() {
        return 0 == this.t ? this.s : (this[0] << 24) >> 24;
      }
      function G() {
        return 0 == this.t ? this.s : (this[0] << 16) >> 16;
      }
      function $(t) {
        return Math.floor((Math.LN2 * this.DB) / Math.log(t));
      }
      function Y() {
        return this.s < 0 ? -1 : this.t <= 0 || (1 == this.t && this[0] <= 0) ? 0 : 1;
      }
      function W(t) {
        if ((null == t && (t = 10), 0 == this.signum() || 2 > t || t > 36)) return "0";
        var e = this.chunkSize(t),
          r = Math.pow(t, e),
          s = c(r),
          n = i(),
          o = i(),
          h = "";
        for (this.divRemTo(s, n, o); n.signum() > 0;)
          ((h = (r + o.intValue()).toString(t).substr(1) + h), n.divRemTo(s, n, o));
        return o.intValue().toString(t) + h;
      }
      function Q(t, i) {
        (this.fromInt(0), null == i && (i = 10));
        for (var r = this.chunkSize(i), s = Math.pow(i, r), n = !1, o = 0, a = 0, u = 0; u < t.length; ++u) {
          var c = h(t, u);
          0 > c
            ? "-" == t.charAt(u) && 0 == this.signum() && (n = !0)
            : ((a = i * a + c), ++o >= r && (this.dMultiply(s), this.dAddOffset(a, 0), (o = 0), (a = 0)));
        }
        (o > 0 && (this.dMultiply(Math.pow(i, o)), this.dAddOffset(a, 0)), n && e.ZERO.subTo(this, this));
      }
      function X(t, i, r) {
        if ("number" == typeof i)
          if (2 > t) this.fromInt(1);
          else
            for (
              this.fromNumber(t, r),
                this.testBit(t - 1) || this.bitwiseTo(e.ONE.shiftLeft(t - 1), he, this),
                this.isEven() && this.dAddOffset(1, 0);
              !this.isProbablePrime(i);
            )
              (this.dAddOffset(2, 0), this.bitLength() > t && this.subTo(e.ONE.shiftLeft(t - 1), this));
        else {
          var s = new Array(),
            n = 7 & t;
          ((s.length = (t >> 3) + 1),
            i.nextBytes(s),
            n > 0 ? (s[0] &= (1 << n) - 1) : (s[0] = 0),
            this.fromString(s, 256));
        }
      }
      function te() {
        var t = this.t,
          e = new Array();
        e[0] = this.s;
        var i,
          r = this.DB - ((t * this.DB) % 8),
          s = 0;
        if (t-- > 0)
          for (
            r < this.DB && (i = this[t] >> r) != (this.s & this.DM) >> r && (e[s++] = i | (this.s << (this.DB - r)));
            t >= 0;
          )
            (8 > r
              ? ((i = (this[t] & ((1 << r) - 1)) << (8 - r)), (i |= this[--t] >> (r += this.DB - 8)))
              : ((i = (this[t] >> (r -= 8)) & 255), 0 >= r && ((r += this.DB), --t)),
              0 != (128 & i) && (i |= -256),
              0 == s && (128 & this.s) != (128 & i) && ++s,
              (s > 0 || i != this.s) && (e[s++] = i));
        return e;
      }
      function ee(t) {
        return 0 == this.compareTo(t);
      }
      function ie(t) {
        return this.compareTo(t) < 0 ? this : t;
      }
      function re(t) {
        return this.compareTo(t) > 0 ? this : t;
      }
      function se(t, e, i) {
        var r,
          s,
          n = Math.min(t.t, this.t);
        for (r = 0; n > r; ++r) i[r] = e(this[r], t[r]);
        if (t.t < this.t) {
          for (s = t.s & this.DM, r = n; r < this.t; ++r) i[r] = e(this[r], s);
          i.t = this.t;
        } else {
          for (s = this.s & this.DM, r = n; r < t.t; ++r) i[r] = e(s, t[r]);
          i.t = t.t;
        }
        ((i.s = e(this.s, t.s)), i.clamp());
      }
      function ne(t, e) {
        return t & e;
      }
      function oe(t) {
        var e = i();
        return (this.bitwiseTo(t, ne, e), e);
      }
      function he(t, e) {
        return t | e;
      }
      function ae(t) {
        var e = i();
        return (this.bitwiseTo(t, he, e), e);
      }
      function ue(t, e) {
        return t ^ e;
      }
      function ce(t) {
        var e = i();
        return (this.bitwiseTo(t, ue, e), e);
      }
      function fe(t, e) {
        return t & ~e;
      }
      function pe(t) {
        var e = i();
        return (this.bitwiseTo(t, fe, e), e);
      }
      function le() {
        for (var t = i(), e = 0; e < this.t; ++e) t[e] = this.DM & ~this[e];
        return ((t.t = this.t), (t.s = ~this.s), t);
      }
      function de(t) {
        var e = i();
        return (0 > t ? this.rShiftTo(-t, e) : this.lShiftTo(t, e), e);
      }
      function ge(t) {
        var e = i();
        return (0 > t ? this.lShiftTo(-t, e) : this.rShiftTo(t, e), e);
      }
      function me(t) {
        if (0 == t) return -1;
        var e = 0;
        return (
          0 == (65535 & t) && ((t >>= 16), (e += 16)),
          0 == (255 & t) && ((t >>= 8), (e += 8)),
          0 == (15 & t) && ((t >>= 4), (e += 4)),
          0 == (3 & t) && ((t >>= 2), (e += 2)),
          0 == (1 & t) && ++e,
          e
        );
      }
      function ye() {
        for (var t = 0; t < this.t; ++t) if (0 != this[t]) return t * this.DB + me(this[t]);
        return this.s < 0 ? this.t * this.DB : -1;
      }
      function ve(t) {
        for (var e = 0; 0 != t;) ((t &= t - 1), ++e);
        return e;
      }
      function be() {
        for (var t = 0, e = this.s & this.DM, i = 0; i < this.t; ++i) t += ve(this[i] ^ e);
        return t;
      }
      function Te(t) {
        var e = Math.floor(t / this.DB);
        return e >= this.t ? 0 != this.s : 0 != (this[e] & (1 << (t % this.DB)));
      }
      function Se(t, i) {
        var r = e.ONE.shiftLeft(t);
        return (this.bitwiseTo(r, i, r), r);
      }
      function Re(t) {
        return this.changeBit(t, he);
      }
      function Ee(t) {
        return this.changeBit(t, fe);
      }
      function De(t) {
        return this.changeBit(t, ue);
      }
      function we(t, e) {
        for (var i = 0, r = 0, s = Math.min(t.t, this.t); s > i;)
          ((r += this[i] + t[i]), (e[i++] = r & this.DM), (r >>= this.DB));
        if (t.t < this.t) {
          for (r += t.s; i < this.t;) ((r += this[i]), (e[i++] = r & this.DM), (r >>= this.DB));
          r += this.s;
        } else {
          for (r += this.s; i < t.t;) ((r += t[i]), (e[i++] = r & this.DM), (r >>= this.DB));
          r += t.s;
        }
        ((e.s = 0 > r ? -1 : 0), r > 0 ? (e[i++] = r) : -1 > r && (e[i++] = this.DV + r), (e.t = i), e.clamp());
      }
      function xe(t) {
        var e = i();
        return (this.addTo(t, e), e);
      }
      function Ae(t) {
        var e = i();
        return (this.subTo(t, e), e);
      }
      function Be(t) {
        var e = i();
        return (this.multiplyTo(t, e), e);
      }
      function Ke() {
        var t = i();
        return (this.squareTo(t), t);
      }
      function Ue(t) {
        var e = i();
        return (this.divRemTo(t, e, null), e);
      }
      function Oe(t) {
        var e = i();
        return (this.divRemTo(t, null, e), e);
      }
      function Ve(t) {
        var e = i(),
          r = i();
        return (this.divRemTo(t, e, r), new Array(e, r));
      }
      function Ne(t) {
        ((this[this.t] = this.am(0, t - 1, this, 0, 0, this.t)), ++this.t, this.clamp());
      }
      function Je(t, e) {
        if (0 != t) {
          for (; this.t <= e;) this[this.t++] = 0;
          for (this[e] += t; this[e] >= this.DV;)
            ((this[e] -= this.DV), ++e >= this.t && (this[this.t++] = 0), ++this[e]);
        }
      }
      function Ie() {}
      function Pe(t) {
        return t;
      }
      function Me(t, e, i) {
        t.multiplyTo(e, i);
      }
      function Le(t, e) {
        t.squareTo(e);
      }
      function qe(t) {
        return this.exp(t, new Ie());
      }
      function He(t, e, i) {
        var r = Math.min(this.t + t.t, e);
        for (i.s = 0, i.t = r; r > 0;) i[--r] = 0;
        var s;
        for (s = i.t - this.t; s > r; ++r) i[r + this.t] = this.am(0, t[r], i, r, 0, this.t);
        for (s = Math.min(t.t, e); s > r; ++r) this.am(0, t[r], i, r, 0, e - r);
        i.clamp();
      }
      function Ce(t, e, i) {
        --e;
        var r = (i.t = this.t + t.t - e);
        for (i.s = 0; --r >= 0;) i[r] = 0;
        for (r = Math.max(e - this.t, 0); r < t.t; ++r)
          i[this.t + r - e] = this.am(e - r, t[r], i, 0, 0, this.t + r - e);
        (i.clamp(), i.drShiftTo(1, i));
      }
      function je(t) {
        ((this.r2 = i()),
          (this.q3 = i()),
          e.ONE.dlShiftTo(2 * t.t, this.r2),
          (this.mu = this.r2.divide(t)),
          (this.m = t));
      }
      function ke(t) {
        if (t.s < 0 || t.t > 2 * this.m.t) return t.mod(this.m);
        if (t.compareTo(this.m) < 0) return t;
        var e = i();
        return (t.copyTo(e), this.reduce(e), e);
      }
      function Fe(t) {
        return t;
      }
      function _e(t) {
        for (
          t.drShiftTo(this.m.t - 1, this.r2),
            t.t > this.m.t + 1 && ((t.t = this.m.t + 1), t.clamp()),
            this.mu.multiplyUpperTo(this.r2, this.m.t + 1, this.q3),
            this.m.multiplyLowerTo(this.q3, this.m.t + 1, this.r2);
          t.compareTo(this.r2) < 0;
        )
          t.dAddOffset(1, this.m.t + 1);
        for (t.subTo(this.r2, t); t.compareTo(this.m) >= 0;) t.subTo(this.m, t);
      }
      function ze(t, e) {
        (t.squareTo(e), this.reduce(e));
      }
      function Ze(t, e, i) {
        (t.multiplyTo(e, i), this.reduce(i));
      }
      function Ge(t, e) {
        var r,
          s,
          n = t.bitLength(),
          o = c(1);
        if (0 >= n) return o;
        ((r = 18 > n ? 1 : 48 > n ? 3 : 144 > n ? 4 : 768 > n ? 5 : 6),
          (s = 8 > n ? new K(e) : e.isEven() ? new je(e) : new P(e)));
        var h = new Array(),
          a = 3,
          u = r - 1,
          f = (1 << r) - 1;
        if (((h[1] = s.convert(this)), r > 1)) {
          var p = i();
          for (s.sqrTo(h[1], p); f >= a;) ((h[a] = i()), s.mulTo(p, h[a - 2], h[a]), (a += 2));
        }
        var l,
          d,
          g = t.t - 1,
          m = !0,
          v = i();
        for (n = y(t[g]) - 1; g >= 0;) {
          for (
            n >= u
              ? (l = (t[g] >> (n - u)) & f)
              : ((l = (t[g] & ((1 << (n + 1)) - 1)) << (u - n)), g > 0 && (l |= t[g - 1] >> (this.DB + n - u))),
              a = r;
            0 == (1 & l);
          )
            ((l >>= 1), --a);
          if (((n -= a) < 0 && ((n += this.DB), --g), m)) (h[l].copyTo(o), (m = !1));
          else {
            for (; a > 1;) (s.sqrTo(o, v), s.sqrTo(v, o), (a -= 2));
            (a > 0 ? s.sqrTo(o, v) : ((d = o), (o = v), (v = d)), s.mulTo(v, h[l], o));
          }
          for (; g >= 0 && 0 == (t[g] & (1 << n));)
            (s.sqrTo(o, v), (d = o), (o = v), (v = d), --n < 0 && ((n = this.DB - 1), --g));
        }
        return s.revert(o);
      }
      function $e(t) {
        var e = this.s < 0 ? this.negate() : this.clone(),
          i = t.s < 0 ? t.negate() : t.clone();
        if (e.compareTo(i) < 0) {
          var r = e;
          ((e = i), (i = r));
        }
        var s = e.getLowestSetBit(),
          n = i.getLowestSetBit();
        if (0 > n) return e;
        for (n > s && (n = s), n > 0 && (e.rShiftTo(n, e), i.rShiftTo(n, i)); e.signum() > 0;)
          ((s = e.getLowestSetBit()) > 0 && e.rShiftTo(s, e),
            (s = i.getLowestSetBit()) > 0 && i.rShiftTo(s, i),
            e.compareTo(i) >= 0 ? (e.subTo(i, e), e.rShiftTo(1, e)) : (i.subTo(e, i), i.rShiftTo(1, i)));
        return (n > 0 && i.lShiftTo(n, i), i);
      }
      function Ye(t) {
        if (0 >= t) return 0;
        var e = this.DV % t,
          i = this.s < 0 ? t - 1 : 0;
        if (this.t > 0)
          if (0 == e) i = this[0] % t;
          else for (var r = this.t - 1; r >= 0; --r) i = (e * i + this[r]) % t;
        return i;
      }
      function We(t) {
        var i = t.isEven();
        if ((this.isEven() && i) || 0 == t.signum()) return e.ZERO;
        for (var r = t.clone(), s = this.clone(), n = c(1), o = c(0), h = c(0), a = c(1); 0 != r.signum();) {
          for (; r.isEven();)
            (r.rShiftTo(1, r),
              i
                ? ((n.isEven() && o.isEven()) || (n.addTo(this, n), o.subTo(t, o)), n.rShiftTo(1, n))
                : o.isEven() || o.subTo(t, o),
              o.rShiftTo(1, o));
          for (; s.isEven();)
            (s.rShiftTo(1, s),
              i
                ? ((h.isEven() && a.isEven()) || (h.addTo(this, h), a.subTo(t, a)), h.rShiftTo(1, h))
                : a.isEven() || a.subTo(t, a),
              a.rShiftTo(1, a));
          r.compareTo(s) >= 0
            ? (r.subTo(s, r), i && n.subTo(h, n), o.subTo(a, o))
            : (s.subTo(r, s), i && h.subTo(n, h), a.subTo(o, a));
        }
        return 0 != s.compareTo(e.ONE)
          ? e.ZERO
          : a.compareTo(t) >= 0
            ? a.subtract(t)
            : a.signum() < 0
              ? (a.addTo(t, a), a.signum() < 0 ? a.add(t) : a)
              : a;
      }
      function Qe(t) {
        var e,
          i = this.abs();
        if (1 == i.t && i[0] <= Oi[Oi.length - 1]) {
          for (e = 0; e < Oi.length; ++e) if (i[0] == Oi[e]) return !0;
          return !1;
        }
        if (i.isEven()) return !1;
        for (e = 1; e < Oi.length;) {
          for (var r = Oi[e], s = e + 1; s < Oi.length && Vi > r;) r *= Oi[s++];
          for (r = i.modInt(r); s > e;) if (r % Oi[e++] == 0) return !1;
        }
        return i.millerRabin(t);
      }
      function Xe(t) {
        var r = this.subtract(e.ONE),
          s = r.getLowestSetBit();
        if (0 >= s) return !1;
        var n = r.shiftRight(s);
        ((t = (t + 1) >> 1), t > Oi.length && (t = Oi.length));
        for (var o = i(), h = 0; t > h; ++h) {
          o.fromInt(Oi[Math.floor(Math.random() * Oi.length)]);
          var a = o.modPow(n, this);
          if (0 != a.compareTo(e.ONE) && 0 != a.compareTo(r)) {
            for (var u = 1; u++ < s && 0 != a.compareTo(r);)
              if (((a = a.modPowInt(2, this)), 0 == a.compareTo(e.ONE))) return !1;
            if (0 != a.compareTo(r)) return !1;
          }
        }
        return !0;
      }
      function ti() {
        ((this.i = 0), (this.j = 0), (this.S = new Array()));
      }
      function ei(t) {
        var e, i, r;
        for (e = 0; 256 > e; ++e) this.S[e] = e;
        for (i = 0, e = 0; 256 > e; ++e)
          ((i = (i + this.S[e] + t[e % t.length]) & 255), (r = this.S[e]), (this.S[e] = this.S[i]), (this.S[i] = r));
        ((this.i = 0), (this.j = 0));
      }
      function ii() {
        var t;
        return (
          (this.i = (this.i + 1) & 255),
          (this.j = (this.j + this.S[this.i]) & 255),
          (t = this.S[this.i]),
          (this.S[this.i] = this.S[this.j]),
          (this.S[this.j] = t),
          this.S[(t + this.S[this.i]) & 255]
        );
      }
      function ri() {
        return new ti();
      }
      function si(t) {
        ((Ji[Ii++] ^= 255 & t),
          (Ji[Ii++] ^= (t >> 8) & 255),
          (Ji[Ii++] ^= (t >> 16) & 255),
          (Ji[Ii++] ^= (t >> 24) & 255),
          Ii >= Pi && (Ii -= Pi));
      }
      function ni() {
        si(new Date().getTime());
      }
      function oi() {
        if (null == Ni) {
          for (ni(), Ni = ri(), Ni.init(Ji), Ii = 0; Ii < Ji.length; ++Ii) Ji[Ii] = 0;
          Ii = 0;
        }
        return Ni.next();
      }
      function hi(t) {
        var e;
        for (e = 0; e < t.length; ++e) t[e] = oi();
      }
      function ai() {}
      function ui(t, i) {
        return new e(t, i);
      }
      function ci(t, i) {
        if (i < t.length + 11) return (console.error("Message too long for RSA"), null);
        for (var r = new Array(), s = t.length - 1; s >= 0 && i > 0;) {
          var n = t.charCodeAt(s--);
          128 > n
            ? (r[--i] = n)
            : n > 127 && 2048 > n
              ? ((r[--i] = (63 & n) | 128), (r[--i] = (n >> 6) | 192))
              : ((r[--i] = (63 & n) | 128), (r[--i] = ((n >> 6) & 63) | 128), (r[--i] = (n >> 12) | 224));
        }
        r[--i] = 0;
        for (var o = new ai(), h = new Array(); i > 2;) {
          for (h[0] = 0; 0 == h[0];) o.nextBytes(h);
          r[--i] = h[0];
        }
        return ((r[--i] = 2), (r[--i] = 0), new e(r));
      }
      function fi() {
        ((this.n = null),
          (this.e = 0),
          (this.d = null),
          (this.p = null),
          (this.q = null),
          (this.dmp1 = null),
          (this.dmq1 = null),
          (this.coeff = null));
      }
      function pi(t, e) {
        null != t && null != e && t.length > 0 && e.length > 0
          ? ((this.n = ui(t, 16)), (this.e = parseInt(e, 16)))
          : console.error("Invalid RSA public key");
      }
      function li(t) {
        return t.modPowInt(this.e, this.n);
      }
      function di(t) {
        var e = ci(t, (this.n.bitLength() + 7) >> 3);
        if (null == e) return null;
        var i = this.doPublic(e);
        if (null == i) return null;
        var r = i.toString(16);
        return 0 == (1 & r.length) ? r : "0" + r;
      }
      function gi(t, e) {
        for (var i = t.toByteArray(), r = 0; r < i.length && 0 == i[r];) ++r;
        if (i.length - r != e - 1 || 2 != i[r]) return null;
        for (++r; 0 != i[r];) if (++r >= i.length) return null;
        for (var s = ""; ++r < i.length;) {
          var n = 255 & i[r];
          128 > n
            ? (s += String.fromCharCode(n))
            : n > 191 && 224 > n
              ? ((s += String.fromCharCode(((31 & n) << 6) | (63 & i[r + 1]))), ++r)
              : ((s += String.fromCharCode(((15 & n) << 12) | ((63 & i[r + 1]) << 6) | (63 & i[r + 2]))), (r += 2));
        }
        return s;
      }
      function mi(t, e, i) {
        null != t && null != e && t.length > 0 && e.length > 0
          ? ((this.n = ui(t, 16)), (this.e = parseInt(e, 16)), (this.d = ui(i, 16)))
          : console.error("Invalid RSA private key");
      }
      function yi(t, e, i, r, s, n, o, h) {
        null != t && null != e && t.length > 0 && e.length > 0
          ? ((this.n = ui(t, 16)),
            (this.e = parseInt(e, 16)),
            (this.d = ui(i, 16)),
            (this.p = ui(r, 16)),
            (this.q = ui(s, 16)),
            (this.dmp1 = ui(n, 16)),
            (this.dmq1 = ui(o, 16)),
            (this.coeff = ui(h, 16)))
          : console.error("Invalid RSA private key");
      }
      function vi(t, i) {
        var r = new ai(),
          s = t >> 1;
        this.e = parseInt(i, 16);
        for (var n = new e(i, 16); ;) {
          for (
            ;
            (this.p = new e(t - s, 1, r)),
              0 != this.p.subtract(e.ONE).gcd(n).compareTo(e.ONE) || !this.p.isProbablePrime(10);
          );
          for (
            ;
            (this.q = new e(s, 1, r)),
              0 != this.q.subtract(e.ONE).gcd(n).compareTo(e.ONE) || !this.q.isProbablePrime(10);
          );
          if (this.p.compareTo(this.q) <= 0) {
            var o = this.p;
            ((this.p = this.q), (this.q = o));
          }
          var h = this.p.subtract(e.ONE),
            a = this.q.subtract(e.ONE),
            u = h.multiply(a);
          if (0 == u.gcd(n).compareTo(e.ONE)) {
            ((this.n = this.p.multiply(this.q)),
              (this.d = n.modInverse(u)),
              (this.dmp1 = this.d.mod(h)),
              (this.dmq1 = this.d.mod(a)),
              (this.coeff = this.q.modInverse(this.p)));
            break;
          }
        }
      }
      function bi(t) {
        if (null == this.p || null == this.q) return t.modPow(this.d, this.n);
        for (
          var e = t.mod(this.p).modPow(this.dmp1, this.p), i = t.mod(this.q).modPow(this.dmq1, this.q);
          e.compareTo(i) < 0;
        )
          e = e.add(this.p);
        return e.subtract(i).multiply(this.coeff).mod(this.p).multiply(this.q).add(i);
      }
      function Ti(t) {
        var e = ui(t, 16),
          i = this.doPrivate(e);
        return null == i ? null : gi(i, (this.n.bitLength() + 7) >> 3);
      }
      function Si(t) {
        var e,
          i,
          r = "";
        for (e = 0; e + 3 <= t.length; e += 3)
          ((i = parseInt(t.substring(e, e + 3), 16)), (r += qi.charAt(i >> 6) + qi.charAt(63 & i)));
        for (
          e + 1 == t.length
            ? ((i = parseInt(t.substring(e, e + 1), 16)), (r += qi.charAt(i << 2)))
            : e + 2 == t.length &&
              ((i = parseInt(t.substring(e, e + 2), 16)), (r += qi.charAt(i >> 2) + qi.charAt((3 & i) << 4)));
          (3 & r.length) > 0;
        )
          r += Hi;
        return r;
      }
      function Ri(t) {
        var e,
          i,
          r = "",
          s = 0;
        for (e = 0; e < t.length && t.charAt(e) != Hi; ++e)
          ((v = qi.indexOf(t.charAt(e))),
            0 > v ||
              (0 == s
                ? ((r += o(v >> 2)), (i = 3 & v), (s = 1))
                : 1 == s
                  ? ((r += o((i << 2) | (v >> 4))), (i = 15 & v), (s = 2))
                  : 2 == s
                    ? ((r += o(i)), (r += o(v >> 2)), (i = 3 & v), (s = 3))
                    : ((r += o((i << 2) | (v >> 4))), (r += o(15 & v)), (s = 0))));
        return (1 == s && (r += o(i << 2)), r);
      }
      var Ei,
        Di = 0xdeadbeefcafe,
        wi = 15715070 == (16777215 & Di);
      (wi && "Microsoft Internet Explorer" == navigator.appName
        ? ((e.prototype.am = s), (Ei = 30))
        : wi && "Netscape" != navigator.appName
          ? ((e.prototype.am = r), (Ei = 26))
          : ((e.prototype.am = n), (Ei = 28)),
        (e.prototype.DB = Ei),
        (e.prototype.DM = (1 << Ei) - 1),
        (e.prototype.DV = 1 << Ei));
      var xi = 52;
      ((e.prototype.FV = Math.pow(2, xi)), (e.prototype.F1 = xi - Ei), (e.prototype.F2 = 2 * Ei - xi));
      var Ai,
        Bi,
        Ki = "0123456789abcdefghijklmnopqrstuvwxyz",
        Ui = new Array();
      for (Ai = "0".charCodeAt(0), Bi = 0; 9 >= Bi; ++Bi) Ui[Ai++] = Bi;
      for (Ai = "a".charCodeAt(0), Bi = 10; 36 > Bi; ++Bi) Ui[Ai++] = Bi;
      for (Ai = "A".charCodeAt(0), Bi = 10; 36 > Bi; ++Bi) Ui[Ai++] = Bi;
      ((K.prototype.convert = U),
        (K.prototype.revert = O),
        (K.prototype.reduce = V),
        (K.prototype.mulTo = N),
        (K.prototype.sqrTo = J),
        (P.prototype.convert = M),
        (P.prototype.revert = L),
        (P.prototype.reduce = q),
        (P.prototype.mulTo = C),
        (P.prototype.sqrTo = H),
        (e.prototype.copyTo = a),
        (e.prototype.fromInt = u),
        (e.prototype.fromString = f),
        (e.prototype.clamp = p),
        (e.prototype.dlShiftTo = T),
        (e.prototype.drShiftTo = S),
        (e.prototype.lShiftTo = R),
        (e.prototype.rShiftTo = E),
        (e.prototype.subTo = D),
        (e.prototype.multiplyTo = w),
        (e.prototype.squareTo = x),
        (e.prototype.divRemTo = A),
        (e.prototype.invDigit = I),
        (e.prototype.isEven = j),
        (e.prototype.exp = k),
        (e.prototype.toString = l),
        (e.prototype.negate = d),
        (e.prototype.abs = g),
        (e.prototype.compareTo = m),
        (e.prototype.bitLength = b),
        (e.prototype.mod = B),
        (e.prototype.modPowInt = F),
        (e.ZERO = c(0)),
        (e.ONE = c(1)),
        (Ie.prototype.convert = Pe),
        (Ie.prototype.revert = Pe),
        (Ie.prototype.mulTo = Me),
        (Ie.prototype.sqrTo = Le),
        (je.prototype.convert = ke),
        (je.prototype.revert = Fe),
        (je.prototype.reduce = _e),
        (je.prototype.mulTo = Ze),
        (je.prototype.sqrTo = ze));
      var Oi = [
          2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107,
          109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193, 197, 199, 211, 223, 227, 229,
          233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307, 311, 313, 317, 331, 337, 347, 349, 353, 359,
          367, 373, 379, 383, 389, 397, 401, 409, 419, 421, 431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491,
          499, 503, 509, 521, 523, 541, 547, 557, 563, 569, 571, 577, 587, 593, 599, 601, 607, 613, 617, 619, 631, 641,
          643, 647, 653, 659, 661, 673, 677, 683, 691, 701, 709, 719, 727, 733, 739, 743, 751, 757, 761, 769, 773, 787,
          797, 809, 811, 821, 823, 827, 829, 839, 853, 857, 859, 863, 877, 881, 883, 887, 907, 911, 919, 929, 937, 941,
          947, 953, 967, 971, 977, 983, 991, 997,
        ],
        Vi = (1 << 26) / Oi[Oi.length - 1];
      ((e.prototype.chunkSize = $),
        (e.prototype.toRadix = W),
        (e.prototype.fromRadix = Q),
        (e.prototype.fromNumber = X),
        (e.prototype.bitwiseTo = se),
        (e.prototype.changeBit = Se),
        (e.prototype.addTo = we),
        (e.prototype.dMultiply = Ne),
        (e.prototype.dAddOffset = Je),
        (e.prototype.multiplyLowerTo = He),
        (e.prototype.multiplyUpperTo = Ce),
        (e.prototype.modInt = Ye),
        (e.prototype.millerRabin = Xe),
        (e.prototype.clone = _),
        (e.prototype.intValue = z),
        (e.prototype.byteValue = Z),
        (e.prototype.shortValue = G),
        (e.prototype.signum = Y),
        (e.prototype.toByteArray = te),
        (e.prototype.equals = ee),
        (e.prototype.min = ie),
        (e.prototype.max = re),
        (e.prototype.and = oe),
        (e.prototype.or = ae),
        (e.prototype.xor = ce),
        (e.prototype.andNot = pe),
        (e.prototype.not = le),
        (e.prototype.shiftLeft = de),
        (e.prototype.shiftRight = ge),
        (e.prototype.getLowestSetBit = ye),
        (e.prototype.bitCount = be),
        (e.prototype.testBit = Te),
        (e.prototype.setBit = Re),
        (e.prototype.clearBit = Ee),
        (e.prototype.flipBit = De),
        (e.prototype.add = xe),
        (e.prototype.subtract = Ae),
        (e.prototype.multiply = Be),
        (e.prototype.divide = Ue),
        (e.prototype.remainder = Oe),
        (e.prototype.divideAndRemainder = Ve),
        (e.prototype.modPow = Ge),
        (e.prototype.modInverse = We),
        (e.prototype.pow = qe),
        (e.prototype.gcd = $e),
        (e.prototype.isProbablePrime = Qe),
        (e.prototype.square = Ke),
        (ti.prototype.init = ei),
        (ti.prototype.next = ii));
      var Ni,
        Ji,
        Ii,
        Pi = 256;
      if (null == Ji) {
        ((Ji = new Array()), (Ii = 0));
        var Mi;
        if ("Netscape" == navigator.appName && navigator.appVersion < "5" && window.crypto && window.crypto.random) {
          var Li = window.crypto.random(32);
          for (Mi = 0; Mi < Li.length; ++Mi) Ji[Ii++] = 255 & Li.charCodeAt(Mi);
        }
        for (; Pi > Ii;) ((Mi = Math.floor(65536 * Math.random())), (Ji[Ii++] = Mi >>> 8), (Ji[Ii++] = 255 & Mi));
        ((Ii = 0), ni());
      }
      ((ai.prototype.nextBytes = hi),
        (fi.prototype.doPublic = li),
        (fi.prototype.setPublic = pi),
        (fi.prototype.encrypt = di),
        (fi.prototype.doPrivate = bi),
        (fi.prototype.setPrivate = mi),
        (fi.prototype.setPrivateEx = yi),
        (fi.prototype.generate = vi),
        (fi.prototype.decrypt = Ti),
        (function () {
          var t = function (t, r, s) {
            var n = new ai(),
              o = t >> 1;
            this.e = parseInt(r, 16);
            var h = new e(r, 16),
              a = this,
              u = function () {
                var r = function () {
                    if (a.p.compareTo(a.q) <= 0) {
                      var t = a.p;
                      ((a.p = a.q), (a.q = t));
                    }
                    var i = a.p.subtract(e.ONE),
                      r = a.q.subtract(e.ONE),
                      n = i.multiply(r);
                    0 == n.gcd(h).compareTo(e.ONE)
                      ? ((a.n = a.p.multiply(a.q)),
                        (a.d = h.modInverse(n)),
                        (a.dmp1 = a.d.mod(i)),
                        (a.dmq1 = a.d.mod(r)),
                        (a.coeff = a.q.modInverse(a.p)),
                        setTimeout(function () {
                          s();
                        }, 0))
                      : setTimeout(u, 0);
                  },
                  c = function () {
                    ((a.q = i()),
                      a.q.fromNumberAsync(o, 1, n, function () {
                        a.q.subtract(e.ONE).gcda(h, function (t) {
                          0 == t.compareTo(e.ONE) && a.q.isProbablePrime(10) ? setTimeout(r, 0) : setTimeout(c, 0);
                        });
                      }));
                  },
                  f = function () {
                    ((a.p = i()),
                      a.p.fromNumberAsync(t - o, 1, n, function () {
                        a.p.subtract(e.ONE).gcda(h, function (t) {
                          0 == t.compareTo(e.ONE) && a.p.isProbablePrime(10) ? setTimeout(c, 0) : setTimeout(f, 0);
                        });
                      }));
                  };
                setTimeout(f, 0);
              };
            setTimeout(u, 0);
          };
          fi.prototype.generateAsync = t;
          var r = function (t, e) {
            var i = this.s < 0 ? this.negate() : this.clone(),
              r = t.s < 0 ? t.negate() : t.clone();
            if (i.compareTo(r) < 0) {
              var s = i;
              ((i = r), (r = s));
            }
            var n = i.getLowestSetBit(),
              o = r.getLowestSetBit();
            if (0 > o) return void e(i);
            (o > n && (o = n), o > 0 && (i.rShiftTo(o, i), r.rShiftTo(o, r)));
            var h = function () {
              ((n = i.getLowestSetBit()) > 0 && i.rShiftTo(n, i),
                (n = r.getLowestSetBit()) > 0 && r.rShiftTo(n, r),
                i.compareTo(r) >= 0 ? (i.subTo(r, i), i.rShiftTo(1, i)) : (r.subTo(i, r), r.rShiftTo(1, r)),
                i.signum() > 0
                  ? setTimeout(h, 0)
                  : (o > 0 && r.lShiftTo(o, r),
                    setTimeout(function () {
                      e(r);
                    }, 0)));
            };
            setTimeout(h, 10);
          };
          e.prototype.gcda = r;
          var s = function (t, i, r, s) {
            if ("number" == typeof i)
              if (2 > t) this.fromInt(1);
              else {
                (this.fromNumber(t, r),
                  this.testBit(t - 1) || this.bitwiseTo(e.ONE.shiftLeft(t - 1), he, this),
                  this.isEven() && this.dAddOffset(1, 0));
                var n = this,
                  o = function () {
                    (n.dAddOffset(2, 0),
                      n.bitLength() > t && n.subTo(e.ONE.shiftLeft(t - 1), n),
                      n.isProbablePrime(i)
                        ? setTimeout(function () {
                            s();
                          }, 0)
                        : setTimeout(o, 0));
                  };
                setTimeout(o, 0);
              }
            else {
              var h = new Array(),
                a = 7 & t;
              ((h.length = (t >> 3) + 1),
                i.nextBytes(h),
                a > 0 ? (h[0] &= (1 << a) - 1) : (h[0] = 0),
                this.fromString(h, 256));
            }
          };
          e.prototype.fromNumberAsync = s;
        })());
      var qi = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
        Hi = "=",
        Ci = Ci || {};
      Ci.env = Ci.env || {};
      var ji = Ci,
        ki = Object.prototype,
        Fi = "[object Function]",
        _i = ["toString", "valueOf"];
      ((Ci.env.parseUA = function (t) {
        var e,
          i = function (t) {
            var e = 0;
            return parseFloat(
              t.replace(/\./g, function () {
                return 1 == e++ ? "" : ".";
              }),
            );
          },
          r = navigator,
          s = {
            ie: 0,
            opera: 0,
            gecko: 0,
            webkit: 0,
            chrome: 0,
            mobile: null,
            air: 0,
            ipad: 0,
            iphone: 0,
            ipod: 0,
            ios: null,
            android: 0,
            webos: 0,
            caja: r && r.cajaVersion,
            secure: !1,
            os: null,
          },
          n = t || (navigator && navigator.userAgent),
          o = window && window.location,
          h = o && o.href;
        return (
          (s.secure = h && 0 === h.toLowerCase().indexOf("https")),
          n &&
            (/windows|win32/i.test(n)
              ? (s.os = "windows")
              : /macintosh/i.test(n)
                ? (s.os = "macintosh")
                : /rhino/i.test(n) && (s.os = "rhino"),
            /KHTML/.test(n) && (s.webkit = 1),
            (e = n.match(/AppleWebKit\/([^\s]*)/)),
            e &&
              e[1] &&
              ((s.webkit = i(e[1])),
              / Mobile\//.test(n)
                ? ((s.mobile = "Apple"),
                  (e = n.match(/OS ([^\s]*)/)),
                  e && e[1] && (e = i(e[1].replace("_", "."))),
                  (s.ios = e),
                  (s.ipad = s.ipod = s.iphone = 0),
                  (e = n.match(/iPad|iPod|iPhone/)),
                  e && e[0] && (s[e[0].toLowerCase()] = s.ios))
                : ((e = n.match(/NokiaN[^\/]*|Android \d\.\d|webOS\/\d\.\d/)),
                  e && (s.mobile = e[0]),
                  /webOS/.test(n) &&
                    ((s.mobile = "WebOS"), (e = n.match(/webOS\/([^\s]*);/)), e && e[1] && (s.webos = i(e[1]))),
                  / Android/.test(n) &&
                    ((s.mobile = "Android"), (e = n.match(/Android ([^\s]*);/)), e && e[1] && (s.android = i(e[1])))),
              (e = n.match(/Chrome\/([^\s]*)/)),
              e && e[1] ? (s.chrome = i(e[1])) : ((e = n.match(/AdobeAIR\/([^\s]*)/)), e && (s.air = e[0]))),
            s.webkit ||
              ((e = n.match(/Opera[\s\/]([^\s]*)/)),
              e && e[1]
                ? ((s.opera = i(e[1])),
                  (e = n.match(/Version\/([^\s]*)/)),
                  e && e[1] && (s.opera = i(e[1])),
                  (e = n.match(/Opera Mini[^;]*/)),
                  e && (s.mobile = e[0]))
                : ((e = n.match(/MSIE\s([^;]*)/)),
                  e && e[1]
                    ? (s.ie = i(e[1]))
                    : ((e = n.match(/Gecko\/([^\s]*)/)),
                      e && ((s.gecko = 1), (e = n.match(/rv:([^\s\)]*)/)), e && e[1] && (s.gecko = i(e[1]))))))),
          s
        );
      }),
        (Ci.env.ua = Ci.env.parseUA()),
        (Ci.isFunction = function (t) {
          return "function" == typeof t || ki.toString.apply(t) === Fi;
        }),
        (Ci._IEEnumFix = Ci.env.ua.ie
          ? function (t, e) {
              var i, r, s;
              for (i = 0; i < _i.length; i += 1)
                ((r = _i[i]), (s = e[r]), ji.isFunction(s) && s != ki[r] && (t[r] = s));
            }
          : function () {}),
        (Ci.extend = function (t, e, i) {
          if (!e || !t) throw new Error("extend failed, please check that all dependencies are included.");
          var r,
            s = function () {};
          if (
            ((s.prototype = e.prototype),
            (t.prototype = new s()),
            (t.prototype.constructor = t),
            (t.superclass = e.prototype),
            e.prototype.constructor == ki.constructor && (e.prototype.constructor = e),
            i)
          ) {
            for (r in i) ji.hasOwnProperty(i, r) && (t.prototype[r] = i[r]);
            ji._IEEnumFix(t.prototype, i);
          }
        }),
        ("undefined" != typeof KJUR && KJUR) || (KJUR = {}),
        ("undefined" != typeof KJUR.asn1 && KJUR.asn1) || (KJUR.asn1 = {}),
        (KJUR.asn1.ASN1Util = new (function () {
          ((this.integerToByteHex = function (t) {
            var e = t.toString(16);
            return (e.length % 2 == 1 && (e = "0" + e), e);
          }),
            (this.bigIntToMinTwosComplementsHex = function (t) {
              var i = t.toString(16);
              if ("-" != i.substr(0, 1)) i.length % 2 == 1 ? (i = "0" + i) : i.match(/^[0-7]/) || (i = "00" + i);
              else {
                var r = i.substr(1),
                  s = r.length;
                s % 2 == 1 ? (s += 1) : i.match(/^[0-7]/) || (s += 2);
                for (var n = "", o = 0; s > o; o++) n += "f";
                var h = new e(n, 16),
                  a = h.xor(t).add(e.ONE);
                i = a.toString(16).replace(/^-/, "");
              }
              return i;
            }),
            (this.getPEMStringFromHex = function (t, e) {
              var i = CryptoJS.enc.Hex.parse(t),
                r = CryptoJS.enc.Base64.stringify(i),
                s = r.replace(/(.{64})/g, "$1\r\n");
              return (
                (s = s.replace(/\r\n$/, "")),
                "-----BEGIN " + e + "-----\r\n" + s + "\r\n-----END " + e + "-----\r\n"
              );
            }));
        })()),
        (KJUR.asn1.ASN1Object = function () {
          var t = "";
          ((this.getLengthHexFromValue = function () {
            if ("undefined" == typeof this.hV || null == this.hV) throw "this.hV is null or undefined.";
            if (this.hV.length % 2 == 1) throw "value hex must be even length: n=" + t.length + ",v=" + this.hV;
            var e = this.hV.length / 2,
              i = e.toString(16);
            if ((i.length % 2 == 1 && (i = "0" + i), 128 > e)) return i;
            var r = i.length / 2;
            if (r > 15) throw "ASN.1 length too long to represent by 8x: n = " + e.toString(16);
            var s = 128 + r;
            return s.toString(16) + i;
          }),
            (this.getEncodedHex = function () {
              return (
                (null == this.hTLV || this.isModified) &&
                  ((this.hV = this.getFreshValueHex()),
                  (this.hL = this.getLengthHexFromValue()),
                  (this.hTLV = this.hT + this.hL + this.hV),
                  (this.isModified = !1)),
                this.hTLV
              );
            }),
            (this.getValueHex = function () {
              return (this.getEncodedHex(), this.hV);
            }),
            (this.getFreshValueHex = function () {
              return "";
            }));
        }),
        (KJUR.asn1.DERAbstractString = function (t) {
          KJUR.asn1.DERAbstractString.superclass.constructor.call(this);
          ((this.getString = function () {
            return this.s;
          }),
            (this.setString = function (t) {
              ((this.hTLV = null), (this.isModified = !0), (this.s = t), (this.hV = stohex(this.s)));
            }),
            (this.setStringHex = function (t) {
              ((this.hTLV = null), (this.isModified = !0), (this.s = null), (this.hV = t));
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.str
                ? this.setString(t.str)
                : "undefined" != typeof t.hex && this.setStringHex(t.hex)));
        }),
        Ci.extend(KJUR.asn1.DERAbstractString, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERAbstractTime = function () {
          KJUR.asn1.DERAbstractTime.superclass.constructor.call(this);
          ((this.localDateToUTC = function (t) {
            utc = t.getTime() + 6e4 * t.getTimezoneOffset();
            var e = new Date(utc);
            return e;
          }),
            (this.formatDate = function (t, e) {
              var i = this.zeroPadding,
                r = this.localDateToUTC(t),
                s = String(r.getFullYear());
              "utc" == e && (s = s.substr(2, 2));
              var n = i(String(r.getMonth() + 1), 2),
                o = i(String(r.getDate()), 2),
                h = i(String(r.getHours()), 2),
                a = i(String(r.getMinutes()), 2),
                u = i(String(r.getSeconds()), 2);
              return s + n + o + h + a + u + "Z";
            }),
            (this.zeroPadding = function (t, e) {
              return t.length >= e ? t : new Array(e - t.length + 1).join("0") + t;
            }),
            (this.getString = function () {
              return this.s;
            }),
            (this.setString = function (t) {
              ((this.hTLV = null), (this.isModified = !0), (this.s = t), (this.hV = stohex(this.s)));
            }),
            (this.setByDateValue = function (t, e, i, r, s, n) {
              var o = new Date(Date.UTC(t, e - 1, i, r, s, n, 0));
              this.setByDate(o);
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }));
        }),
        Ci.extend(KJUR.asn1.DERAbstractTime, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERAbstractStructured = function (t) {
          KJUR.asn1.DERAbstractString.superclass.constructor.call(this);
          ((this.setByASN1ObjectArray = function (t) {
            ((this.hTLV = null), (this.isModified = !0), (this.asn1Array = t));
          }),
            (this.appendASN1Object = function (t) {
              ((this.hTLV = null), (this.isModified = !0), this.asn1Array.push(t));
            }),
            (this.asn1Array = new Array()),
            "undefined" != typeof t && "undefined" != typeof t.array && (this.asn1Array = t.array));
        }),
        Ci.extend(KJUR.asn1.DERAbstractStructured, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERBoolean = function () {
          (KJUR.asn1.DERBoolean.superclass.constructor.call(this), (this.hT = "01"), (this.hTLV = "0101ff"));
        }),
        Ci.extend(KJUR.asn1.DERBoolean, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERInteger = function (t) {
          (KJUR.asn1.DERInteger.superclass.constructor.call(this),
            (this.hT = "02"),
            (this.setByBigInteger = function (t) {
              ((this.hTLV = null),
                (this.isModified = !0),
                (this.hV = KJUR.asn1.ASN1Util.bigIntToMinTwosComplementsHex(t)));
            }),
            (this.setByInteger = function (t) {
              var i = new e(String(t), 10);
              this.setByBigInteger(i);
            }),
            (this.setValueHex = function (t) {
              this.hV = t;
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.bigint
                ? this.setByBigInteger(t.bigint)
                : "undefined" != typeof t["int"]
                  ? this.setByInteger(t["int"])
                  : "undefined" != typeof t.hex && this.setValueHex(t.hex)));
        }),
        Ci.extend(KJUR.asn1.DERInteger, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERBitString = function (t) {
          (KJUR.asn1.DERBitString.superclass.constructor.call(this),
            (this.hT = "03"),
            (this.setHexValueIncludingUnusedBits = function (t) {
              ((this.hTLV = null), (this.isModified = !0), (this.hV = t));
            }),
            (this.setUnusedBitsAndHexValue = function (t, e) {
              if (0 > t || t > 7) throw "unused bits shall be from 0 to 7: u = " + t;
              var i = "0" + t;
              ((this.hTLV = null), (this.isModified = !0), (this.hV = i + e));
            }),
            (this.setByBinaryString = function (t) {
              t = t.replace(/0+$/, "");
              var e = 8 - (t.length % 8);
              8 == e && (e = 0);
              for (var i = 0; e >= i; i++) t += "0";
              for (var r = "", i = 0; i < t.length - 1; i += 8) {
                var s = t.substr(i, 8),
                  n = parseInt(s, 2).toString(16);
                (1 == n.length && (n = "0" + n), (r += n));
              }
              ((this.hTLV = null), (this.isModified = !0), (this.hV = "0" + e + r));
            }),
            (this.setByBooleanArray = function (t) {
              for (var e = "", i = 0; i < t.length; i++) e += 1 == t[i] ? "1" : "0";
              this.setByBinaryString(e);
            }),
            (this.newFalseArray = function (t) {
              for (var e = new Array(t), i = 0; t > i; i++) e[i] = !1;
              return e;
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.hex
                ? this.setHexValueIncludingUnusedBits(t.hex)
                : "undefined" != typeof t.bin
                  ? this.setByBinaryString(t.bin)
                  : "undefined" != typeof t.array && this.setByBooleanArray(t.array)));
        }),
        Ci.extend(KJUR.asn1.DERBitString, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DEROctetString = function (t) {
          (KJUR.asn1.DEROctetString.superclass.constructor.call(this, t), (this.hT = "04"));
        }),
        Ci.extend(KJUR.asn1.DEROctetString, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERNull = function () {
          (KJUR.asn1.DERNull.superclass.constructor.call(this), (this.hT = "05"), (this.hTLV = "0500"));
        }),
        Ci.extend(KJUR.asn1.DERNull, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERObjectIdentifier = function (t) {
          var i = function (t) {
              var e = t.toString(16);
              return (1 == e.length && (e = "0" + e), e);
            },
            r = function (t) {
              var r = "",
                s = new e(t, 10),
                n = s.toString(2),
                o = 7 - (n.length % 7);
              7 == o && (o = 0);
              for (var h = "", a = 0; o > a; a++) h += "0";
              n = h + n;
              for (var a = 0; a < n.length - 1; a += 7) {
                var u = n.substr(a, 7);
                (a != n.length - 7 && (u = "1" + u), (r += i(parseInt(u, 2))));
              }
              return r;
            };
          (KJUR.asn1.DERObjectIdentifier.superclass.constructor.call(this),
            (this.hT = "06"),
            (this.setValueHex = function (t) {
              ((this.hTLV = null), (this.isModified = !0), (this.s = null), (this.hV = t));
            }),
            (this.setValueOidString = function (t) {
              if (!t.match(/^[0-9.]+$/)) throw "malformed oid string: " + t;
              var e = "",
                s = t.split("."),
                n = 40 * parseInt(s[0]) + parseInt(s[1]);
              ((e += i(n)), s.splice(0, 2));
              for (var o = 0; o < s.length; o++) e += r(s[o]);
              ((this.hTLV = null), (this.isModified = !0), (this.s = null), (this.hV = e));
            }),
            (this.setValueName = function (t) {
              if ("undefined" == typeof KJUR.asn1.x509.OID.name2oidList[t])
                throw "DERObjectIdentifier oidName undefined: " + t;
              var e = KJUR.asn1.x509.OID.name2oidList[t];
              this.setValueOidString(e);
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.oid
                ? this.setValueOidString(t.oid)
                : "undefined" != typeof t.hex
                  ? this.setValueHex(t.hex)
                  : "undefined" != typeof t.name && this.setValueName(t.name)));
        }),
        Ci.extend(KJUR.asn1.DERObjectIdentifier, KJUR.asn1.ASN1Object),
        (KJUR.asn1.DERUTF8String = function (t) {
          (KJUR.asn1.DERUTF8String.superclass.constructor.call(this, t), (this.hT = "0c"));
        }),
        Ci.extend(KJUR.asn1.DERUTF8String, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERNumericString = function (t) {
          (KJUR.asn1.DERNumericString.superclass.constructor.call(this, t), (this.hT = "12"));
        }),
        Ci.extend(KJUR.asn1.DERNumericString, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERPrintableString = function (t) {
          (KJUR.asn1.DERPrintableString.superclass.constructor.call(this, t), (this.hT = "13"));
        }),
        Ci.extend(KJUR.asn1.DERPrintableString, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERTeletexString = function (t) {
          (KJUR.asn1.DERTeletexString.superclass.constructor.call(this, t), (this.hT = "14"));
        }),
        Ci.extend(KJUR.asn1.DERTeletexString, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERIA5String = function (t) {
          (KJUR.asn1.DERIA5String.superclass.constructor.call(this, t), (this.hT = "16"));
        }),
        Ci.extend(KJUR.asn1.DERIA5String, KJUR.asn1.DERAbstractString),
        (KJUR.asn1.DERUTCTime = function (t) {
          (KJUR.asn1.DERUTCTime.superclass.constructor.call(this, t),
            (this.hT = "17"),
            (this.setByDate = function (t) {
              ((this.hTLV = null),
                (this.isModified = !0),
                (this.date = t),
                (this.s = this.formatDate(this.date, "utc")),
                (this.hV = stohex(this.s)));
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.str
                ? this.setString(t.str)
                : "undefined" != typeof t.hex
                  ? this.setStringHex(t.hex)
                  : "undefined" != typeof t.date && this.setByDate(t.date)));
        }),
        Ci.extend(KJUR.asn1.DERUTCTime, KJUR.asn1.DERAbstractTime),
        (KJUR.asn1.DERGeneralizedTime = function (t) {
          (KJUR.asn1.DERGeneralizedTime.superclass.constructor.call(this, t),
            (this.hT = "18"),
            (this.setByDate = function (t) {
              ((this.hTLV = null),
                (this.isModified = !0),
                (this.date = t),
                (this.s = this.formatDate(this.date, "gen")),
                (this.hV = stohex(this.s)));
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.str
                ? this.setString(t.str)
                : "undefined" != typeof t.hex
                  ? this.setStringHex(t.hex)
                  : "undefined" != typeof t.date && this.setByDate(t.date)));
        }),
        Ci.extend(KJUR.asn1.DERGeneralizedTime, KJUR.asn1.DERAbstractTime),
        (KJUR.asn1.DERSequence = function (t) {
          (KJUR.asn1.DERSequence.superclass.constructor.call(this, t),
            (this.hT = "30"),
            (this.getFreshValueHex = function () {
              for (var t = "", e = 0; e < this.asn1Array.length; e++) {
                var i = this.asn1Array[e];
                t += i.getEncodedHex();
              }
              return ((this.hV = t), this.hV);
            }));
        }),
        Ci.extend(KJUR.asn1.DERSequence, KJUR.asn1.DERAbstractStructured),
        (KJUR.asn1.DERSet = function (t) {
          (KJUR.asn1.DERSet.superclass.constructor.call(this, t),
            (this.hT = "31"),
            (this.getFreshValueHex = function () {
              for (var t = new Array(), e = 0; e < this.asn1Array.length; e++) {
                var i = this.asn1Array[e];
                t.push(i.getEncodedHex());
              }
              return (t.sort(), (this.hV = t.join("")), this.hV);
            }));
        }),
        Ci.extend(KJUR.asn1.DERSet, KJUR.asn1.DERAbstractStructured),
        (KJUR.asn1.DERTaggedObject = function (t) {
          (KJUR.asn1.DERTaggedObject.superclass.constructor.call(this),
            (this.hT = "a0"),
            (this.hV = ""),
            (this.isExplicit = !0),
            (this.asn1Object = null),
            (this.setASN1Object = function (t, e, i) {
              ((this.hT = e),
                (this.isExplicit = t),
                (this.asn1Object = i),
                this.isExplicit
                  ? ((this.hV = this.asn1Object.getEncodedHex()), (this.hTLV = null), (this.isModified = !0))
                  : ((this.hV = null),
                    (this.hTLV = i.getEncodedHex()),
                    (this.hTLV = this.hTLV.replace(/^../, e)),
                    (this.isModified = !1)));
            }),
            (this.getFreshValueHex = function () {
              return this.hV;
            }),
            "undefined" != typeof t &&
              ("undefined" != typeof t.tag && (this.hT = t.tag),
              "undefined" != typeof t.explicit && (this.isExplicit = t.explicit),
              "undefined" != typeof t.obj &&
                ((this.asn1Object = t.obj), this.setASN1Object(this.isExplicit, this.hT, this.asn1Object))));
        }),
        Ci.extend(KJUR.asn1.DERTaggedObject, KJUR.asn1.ASN1Object),
        (function (t) {
          "use strict";
          var e,
            i = {};
          ((i.decode = function (i) {
            var r;
            if (e === t) {
              var s = "0123456789ABCDEF",
                n = " \f\n\r	 \u2028\u2029";
              for (e = [], r = 0; 16 > r; ++r) e[s.charAt(r)] = r;
              for (s = s.toLowerCase(), r = 10; 16 > r; ++r) e[s.charAt(r)] = r;
              for (r = 0; r < n.length; ++r) e[n.charAt(r)] = -1;
            }
            var o = [],
              h = 0,
              a = 0;
            for (r = 0; r < i.length; ++r) {
              var u = i.charAt(r);
              if ("=" == u) break;
              if (((u = e[u]), -1 != u)) {
                if (u === t) throw "Illegal character at offset " + r;
                ((h |= u), ++a >= 2 ? ((o[o.length] = h), (h = 0), (a = 0)) : (h <<= 4));
              }
            }
            if (a) throw "Hex encoding incomplete: 4 bits missing";
            return o;
          }),
            (window.Hex = i));
        })(),
        (function (t) {
          "use strict";
          var e,
            i = {};
          ((i.decode = function (i) {
            var r;
            if (e === t) {
              var s = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
                n = "= \f\n\r	 \u2028\u2029";
              for (e = [], r = 0; 64 > r; ++r) e[s.charAt(r)] = r;
              for (r = 0; r < n.length; ++r) e[n.charAt(r)] = -1;
            }
            var o = [],
              h = 0,
              a = 0;
            for (r = 0; r < i.length; ++r) {
              var u = i.charAt(r);
              if ("=" == u) break;
              if (((u = e[u]), -1 != u)) {
                if (u === t) throw "Illegal character at offset " + r;
                ((h |= u),
                  ++a >= 4
                    ? ((o[o.length] = h >> 16),
                      (o[o.length] = (h >> 8) & 255),
                      (o[o.length] = 255 & h),
                      (h = 0),
                      (a = 0))
                    : (h <<= 6));
              }
            }
            switch (a) {
              case 1:
                throw "Base64 encoding incomplete: at least 2 bits missing";
              case 2:
                o[o.length] = h >> 10;
                break;
              case 3:
                ((o[o.length] = h >> 16), (o[o.length] = (h >> 8) & 255));
            }
            return o;
          }),
            (i.re =
              /-----BEGIN [^-]+-----([A-Za-z0-9+\/=\s]+)-----END [^-]+-----|begin-base64[^\n]+\n([A-Za-z0-9+\/=\s]+)====/),
            (i.unarmor = function (t) {
              var e = i.re.exec(t);
              if (e)
                if (e[1]) t = e[1];
                else {
                  if (!e[2]) throw "RegExp out of sync";
                  t = e[2];
                }
              return i.decode(t);
            }),
            (window.Base64 = i));
        })(),
        (function (t) {
          "use strict";
          function e(t, i) {
            t instanceof e ? ((this.enc = t.enc), (this.pos = t.pos)) : ((this.enc = t), (this.pos = i));
          }
          function i(t, e, i, r, s) {
            ((this.stream = t), (this.header = e), (this.length = i), (this.tag = r), (this.sub = s));
          }
          var r = 100,
            s = "…",
            n = {
              tag: function (t, e) {
                var i = document.createElement(t);
                return ((i.className = e), i);
              },
              text: function (t) {
                return document.createTextNode(t);
              },
            };
          ((e.prototype.get = function (e) {
            if ((e === t && (e = this.pos++), e >= this.enc.length))
              throw "Requesting byte offset " + e + " on a stream of length " + this.enc.length;
            return this.enc[e];
          }),
            (e.prototype.hexDigits = "0123456789ABCDEF"),
            (e.prototype.hexByte = function (t) {
              return this.hexDigits.charAt((t >> 4) & 15) + this.hexDigits.charAt(15 & t);
            }),
            (e.prototype.hexDump = function (t, e, i) {
              for (var r = "", s = t; e > s; ++s)
                if (((r += this.hexByte(this.get(s))), i !== !0))
                  switch (15 & s) {
                    case 7:
                      r += "  ";
                      break;
                    case 15:
                      r += "\n";
                      break;
                    default:
                      r += " ";
                  }
              return r;
            }),
            (e.prototype.parseStringISO = function (t, e) {
              for (var i = "", r = t; e > r; ++r) i += String.fromCharCode(this.get(r));
              return i;
            }),
            (e.prototype.parseStringUTF = function (t, e) {
              for (var i = "", r = t; e > r;) {
                var s = this.get(r++);
                i += String.fromCharCode(
                  128 > s
                    ? s
                    : s > 191 && 224 > s
                      ? ((31 & s) << 6) | (63 & this.get(r++))
                      : ((15 & s) << 12) | ((63 & this.get(r++)) << 6) | (63 & this.get(r++)),
                );
              }
              return i;
            }),
            (e.prototype.parseStringBMP = function (t, e) {
              for (var i = "", r = t; e > r; r += 2) {
                var s = this.get(r),
                  n = this.get(r + 1);
                i += String.fromCharCode((s << 8) + n);
              }
              return i;
            }),
            (e.prototype.reTime =
              /^((?:1[89]|2\d)?\d\d)(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])([01]\d|2[0-3])(?:([0-5]\d)(?:([0-5]\d)(?:[.,](\d{1,3}))?)?)?(Z|[-+](?:[0]\d|1[0-2])([0-5]\d)?)?$/),
            (e.prototype.parseTime = function (t, e) {
              var i = this.parseStringISO(t, e),
                r = this.reTime.exec(i);
              return r
                ? ((i = r[1] + "-" + r[2] + "-" + r[3] + " " + r[4]),
                  r[5] && ((i += ":" + r[5]), r[6] && ((i += ":" + r[6]), r[7] && (i += "." + r[7]))),
                  r[8] && ((i += " UTC"), "Z" != r[8] && ((i += r[8]), r[9] && (i += ":" + r[9]))),
                  i)
                : "Unrecognized time: " + i;
            }),
            (e.prototype.parseInteger = function (t, e) {
              var i = e - t;
              if (i > 4) {
                i <<= 3;
                var r = this.get(t);
                if (0 === r) i -= 8;
                else for (; 128 > r;) ((r <<= 1), --i);
                return "(" + i + " bit)";
              }
              for (var s = 0, n = t; e > n; ++n) s = (s << 8) | this.get(n);
              return s;
            }),
            (e.prototype.parseBitString = function (t, e) {
              var i = this.get(t),
                r = ((e - t - 1) << 3) - i,
                s = "(" + r + " bit)";
              if (20 >= r) {
                var n = i;
                s += " ";
                for (var o = e - 1; o > t; --o) {
                  for (var h = this.get(o), a = n; 8 > a; ++a) s += (h >> a) & 1 ? "1" : "0";
                  n = 0;
                }
              }
              return s;
            }),
            (e.prototype.parseOctetString = function (t, e) {
              var i = e - t,
                n = "(" + i + " byte) ";
              i > r && (e = t + r);
              for (var o = t; e > o; ++o) n += this.hexByte(this.get(o));
              return (i > r && (n += s), n);
            }),
            (e.prototype.parseOID = function (t, e) {
              for (var i = "", r = 0, s = 0, n = t; e > n; ++n) {
                var o = this.get(n);
                if (((r = (r << 7) | (127 & o)), (s += 7), !(128 & o))) {
                  if ("" === i) {
                    var h = 80 > r ? (40 > r ? 0 : 1) : 2;
                    i = h + "." + (r - 40 * h);
                  } else i += "." + (s >= 31 ? "bigint" : r);
                  r = s = 0;
                }
              }
              return i;
            }),
            (i.prototype.typeName = function () {
              if (this.tag === t) return "unknown";
              var e = this.tag >> 6,
                i = ((this.tag >> 5) & 1, 31 & this.tag);
              switch (e) {
                case 0:
                  switch (i) {
                    case 0:
                      return "EOC";
                    case 1:
                      return "BOOLEAN";
                    case 2:
                      return "INTEGER";
                    case 3:
                      return "BIT_STRING";
                    case 4:
                      return "OCTET_STRING";
                    case 5:
                      return "NULL";
                    case 6:
                      return "OBJECT_IDENTIFIER";
                    case 7:
                      return "ObjectDescriptor";
                    case 8:
                      return "EXTERNAL";
                    case 9:
                      return "REAL";
                    case 10:
                      return "ENUMERATED";
                    case 11:
                      return "EMBEDDED_PDV";
                    case 12:
                      return "UTF8String";
                    case 16:
                      return "SEQUENCE";
                    case 17:
                      return "SET";
                    case 18:
                      return "NumericString";
                    case 19:
                      return "PrintableString";
                    case 20:
                      return "TeletexString";
                    case 21:
                      return "VideotexString";
                    case 22:
                      return "IA5String";
                    case 23:
                      return "UTCTime";
                    case 24:
                      return "GeneralizedTime";
                    case 25:
                      return "GraphicString";
                    case 26:
                      return "VisibleString";
                    case 27:
                      return "GeneralString";
                    case 28:
                      return "UniversalString";
                    case 30:
                      return "BMPString";
                    default:
                      return "Universal_" + i.toString(16);
                  }
                case 1:
                  return "Application_" + i.toString(16);
                case 2:
                  return "[" + i + "]";
                case 3:
                  return "Private_" + i.toString(16);
              }
            }),
            (i.prototype.reSeemsASCII = /^[ -~]+$/),
            (i.prototype.content = function () {
              if (this.tag === t) return null;
              var e = this.tag >> 6,
                i = 31 & this.tag,
                n = this.posContent(),
                o = Math.abs(this.length);
              if (0 !== e) {
                if (null !== this.sub) return "(" + this.sub.length + " elem)";
                var h = this.stream.parseStringISO(n, n + Math.min(o, r));
                return this.reSeemsASCII.test(h)
                  ? h.substring(0, 2 * r) + (h.length > 2 * r ? s : "")
                  : this.stream.parseOctetString(n, n + o);
              }
              switch (i) {
                case 1:
                  return 0 === this.stream.get(n) ? "false" : "true";
                case 2:
                  return this.stream.parseInteger(n, n + o);
                case 3:
                  return this.sub ? "(" + this.sub.length + " elem)" : this.stream.parseBitString(n, n + o);
                case 4:
                  return this.sub ? "(" + this.sub.length + " elem)" : this.stream.parseOctetString(n, n + o);
                case 6:
                  return this.stream.parseOID(n, n + o);
                case 16:
                case 17:
                  return "(" + this.sub.length + " elem)";
                case 12:
                  return this.stream.parseStringUTF(n, n + o);
                case 18:
                case 19:
                case 20:
                case 21:
                case 22:
                case 26:
                  return this.stream.parseStringISO(n, n + o);
                case 30:
                  return this.stream.parseStringBMP(n, n + o);
                case 23:
                case 24:
                  return this.stream.parseTime(n, n + o);
              }
              return null;
            }),
            (i.prototype.toString = function () {
              return (
                this.typeName() +
                "@" +
                this.stream.pos +
                "[header:" +
                this.header +
                ",length:" +
                this.length +
                ",sub:" +
                (null === this.sub ? "null" : this.sub.length) +
                "]"
              );
            }),
            (i.prototype.print = function (e) {
              if ((e === t && (e = ""), document.writeln(e + this), null !== this.sub)) {
                e += "  ";
                for (var i = 0, r = this.sub.length; r > i; ++i) this.sub[i].print(e);
              }
            }),
            (i.prototype.toPrettyString = function (e) {
              e === t && (e = "");
              var i = e + this.typeName() + " @" + this.stream.pos;
              if (
                (this.length >= 0 && (i += "+"),
                (i += this.length),
                32 & this.tag
                  ? (i += " (constructed)")
                  : (3 != this.tag && 4 != this.tag) || null === this.sub || (i += " (encapsulates)"),
                (i += "\n"),
                null !== this.sub)
              ) {
                e += "  ";
                for (var r = 0, s = this.sub.length; s > r; ++r) i += this.sub[r].toPrettyString(e);
              }
              return i;
            }),
            (i.prototype.toDOM = function () {
              var t = n.tag("div", "node");
              t.asn1 = this;
              var e = n.tag("div", "head"),
                i = this.typeName().replace(/_/g, " ");
              e.innerHTML = i;
              var r = this.content();
              if (null !== r) {
                r = String(r).replace(/</g, "&lt;");
                var s = n.tag("span", "preview");
                (s.appendChild(n.text(r)), e.appendChild(s));
              }
              (t.appendChild(e), (this.node = t), (this.head = e));
              var o = n.tag("div", "value");
              if (
                ((i = "Offset: " + this.stream.pos + "<br/>"),
                (i += "Length: " + this.header + "+"),
                (i += this.length >= 0 ? this.length : -this.length + " (undefined)"),
                32 & this.tag
                  ? (i += "<br/>(constructed)")
                  : (3 != this.tag && 4 != this.tag) || null === this.sub || (i += "<br/>(encapsulates)"),
                null !== r && ((i += "<br/>Value:<br/><b>" + r + "</b>"), "object" == typeof oids && 6 == this.tag))
              ) {
                var h = oids[r];
                h && (h.d && (i += "<br/>" + h.d), h.c && (i += "<br/>" + h.c), h.w && (i += "<br/>(warning!)"));
              }
              ((o.innerHTML = i), t.appendChild(o));
              var a = n.tag("div", "sub");
              if (null !== this.sub)
                for (var u = 0, c = this.sub.length; c > u; ++u) a.appendChild(this.sub[u].toDOM());
              return (
                t.appendChild(a),
                (e.onclick = function () {
                  t.className = "node collapsed" == t.className ? "node" : "node collapsed";
                }),
                t
              );
            }),
            (i.prototype.posStart = function () {
              return this.stream.pos;
            }),
            (i.prototype.posContent = function () {
              return this.stream.pos + this.header;
            }),
            (i.prototype.posEnd = function () {
              return this.stream.pos + this.header + Math.abs(this.length);
            }),
            (i.prototype.fakeHover = function (t) {
              ((this.node.className += " hover"), t && (this.head.className += " hover"));
            }),
            (i.prototype.fakeOut = function (t) {
              var e = / ?hover/;
              ((this.node.className = this.node.className.replace(e, "")),
                t && (this.head.className = this.head.className.replace(e, "")));
            }),
            (i.prototype.toHexDOM_sub = function (t, e, i, r, s) {
              if (!(r >= s)) {
                var o = n.tag("span", e);
                (o.appendChild(n.text(i.hexDump(r, s))), t.appendChild(o));
              }
            }),
            (i.prototype.toHexDOM = function (e) {
              var i = n.tag("span", "hex");
              if (
                (e === t && (e = i),
                (this.head.hexNode = i),
                (this.head.onmouseover = function () {
                  this.hexNode.className = "hexCurrent";
                }),
                (this.head.onmouseout = function () {
                  this.hexNode.className = "hex";
                }),
                (i.asn1 = this),
                (i.onmouseover = function () {
                  var t = !e.selected;
                  (t && ((e.selected = this.asn1), (this.className = "hexCurrent")), this.asn1.fakeHover(t));
                }),
                (i.onmouseout = function () {
                  var t = e.selected == this.asn1;
                  (this.asn1.fakeOut(t), t && ((e.selected = null), (this.className = "hex")));
                }),
                this.toHexDOM_sub(i, "tag", this.stream, this.posStart(), this.posStart() + 1),
                this.toHexDOM_sub(
                  i,
                  this.length >= 0 ? "dlen" : "ulen",
                  this.stream,
                  this.posStart() + 1,
                  this.posContent(),
                ),
                null === this.sub)
              )
                i.appendChild(n.text(this.stream.hexDump(this.posContent(), this.posEnd())));
              else if (this.sub.length > 0) {
                var r = this.sub[0],
                  s = this.sub[this.sub.length - 1];
                this.toHexDOM_sub(i, "intro", this.stream, this.posContent(), r.posStart());
                for (var o = 0, h = this.sub.length; h > o; ++o) i.appendChild(this.sub[o].toHexDOM(e));
                this.toHexDOM_sub(i, "outro", this.stream, s.posEnd(), this.posEnd());
              }
              return i;
            }),
            (i.prototype.toHexString = function () {
              return this.stream.hexDump(this.posStart(), this.posEnd(), !0);
            }),
            (i.decodeLength = function (t) {
              var e = t.get(),
                i = 127 & e;
              if (i == e) return i;
              if (i > 3) throw "Length over 24 bits not supported at position " + (t.pos - 1);
              if (0 === i) return -1;
              e = 0;
              for (var r = 0; i > r; ++r) e = (e << 8) | t.get();
              return e;
            }),
            (i.hasContent = function (t, r, s) {
              if (32 & t) return !0;
              if (3 > t || t > 4) return !1;
              var n = new e(s);
              3 == t && n.get();
              var o = n.get();
              if ((o >> 6) & 1) return !1;
              try {
                var h = i.decodeLength(n);
                return n.pos - s.pos + h == r;
              } catch (a) {
                return !1;
              }
            }),
            (i.decode = function (t) {
              t instanceof e || (t = new e(t, 0));
              var r = new e(t),
                s = t.get(),
                n = i.decodeLength(t),
                o = t.pos - r.pos,
                h = null;
              if (i.hasContent(s, n, t)) {
                var a = t.pos;
                if ((3 == s && t.get(), (h = []), n >= 0)) {
                  for (var u = a + n; t.pos < u;) h[h.length] = i.decode(t);
                  if (t.pos != u) throw "Content size is not correct for container starting at offset " + a;
                } else
                  try {
                    for (;;) {
                      var c = i.decode(t);
                      if (0 === c.tag) break;
                      h[h.length] = c;
                    }
                    n = a - t.pos;
                  } catch (f) {
                    throw "Exception while decoding undefined length content: " + f;
                  }
              } else t.pos += n;
              return new i(r, o, n, s, h);
            }),
            (i.test = function () {
              for (
                var t = [
                    { value: [39], expected: 39 },
                    { value: [129, 201], expected: 201 },
                    { value: [131, 254, 220, 186], expected: 16702650 },
                  ],
                  r = 0,
                  s = t.length;
                s > r;
                ++r
              ) {
                var n = new e(t[r].value, 0),
                  o = i.decodeLength(n);
                o != t[r].expected &&
                  document.write("In test[" + r + "] expected " + t[r].expected + " got " + o + "\n");
              }
            }),
            (window.ASN1 = i));
        })(),
        (ASN1.prototype.getHexStringValue = function () {
          var t = this.toHexString(),
            e = 2 * this.header,
            i = 2 * this.length;
          return t.substr(e, i);
        }),
        (fi.prototype.parseKey = function (t) {
          try {
            var e = /^\s*(?:[0-9A-Fa-f][0-9A-Fa-f]\s*)+$/,
              i = e.test(t) ? Hex.decode(t) : Base64.unarmor(t),
              r = ASN1.decode(i);
            if (9 === r.sub.length) {
              var s = r.sub[1].getHexStringValue();
              this.n = ui(s, 16);
              var n = r.sub[2].getHexStringValue();
              this.e = parseInt(n, 16);
              var o = r.sub[3].getHexStringValue();
              this.d = ui(o, 16);
              var h = r.sub[4].getHexStringValue();
              this.p = ui(h, 16);
              var a = r.sub[5].getHexStringValue();
              this.q = ui(a, 16);
              var u = r.sub[6].getHexStringValue();
              this.dmp1 = ui(u, 16);
              var c = r.sub[7].getHexStringValue();
              this.dmq1 = ui(c, 16);
              var f = r.sub[8].getHexStringValue();
              this.coeff = ui(f, 16);
            } else {
              if (2 !== r.sub.length) return !1;
              var p = r.sub[1],
                l = p.sub[0],
                s = l.sub[0].getHexStringValue();
              this.n = ui(s, 16);
              var n = l.sub[1].getHexStringValue();
              this.e = parseInt(n, 16);
            }
            return !0;
          } catch (d) {
            return !1;
          }
        }),
        (fi.prototype.getPrivateBaseKey = function () {
          var t = {
              array: [
                new KJUR.asn1.DERInteger({ int: 0 }),
                new KJUR.asn1.DERInteger({ bigint: this.n }),
                new KJUR.asn1.DERInteger({ int: this.e }),
                new KJUR.asn1.DERInteger({ bigint: this.d }),
                new KJUR.asn1.DERInteger({ bigint: this.p }),
                new KJUR.asn1.DERInteger({ bigint: this.q }),
                new KJUR.asn1.DERInteger({ bigint: this.dmp1 }),
                new KJUR.asn1.DERInteger({ bigint: this.dmq1 }),
                new KJUR.asn1.DERInteger({ bigint: this.coeff }),
              ],
            },
            e = new KJUR.asn1.DERSequence(t);
          return e.getEncodedHex();
        }),
        (fi.prototype.getPrivateBaseKeyB64 = function () {
          return Si(this.getPrivateBaseKey());
        }),
        (fi.prototype.getPublicBaseKey = function () {
          var t = {
              array: [new KJUR.asn1.DERObjectIdentifier({ oid: "1.2.840.113549.1.1.1" }), new KJUR.asn1.DERNull()],
            },
            e = new KJUR.asn1.DERSequence(t);
          t = { array: [new KJUR.asn1.DERInteger({ bigint: this.n }), new KJUR.asn1.DERInteger({ int: this.e })] };
          var i = new KJUR.asn1.DERSequence(t);
          t = { hex: "00" + i.getEncodedHex() };
          var r = new KJUR.asn1.DERBitString(t);
          t = { array: [e, r] };
          var s = new KJUR.asn1.DERSequence(t);
          return s.getEncodedHex();
        }),
        (fi.prototype.getPublicBaseKeyB64 = function () {
          return Si(this.getPublicBaseKey());
        }),
        (fi.prototype.wordwrap = function (t, e) {
          if (((e = e || 64), !t)) return t;
          var i = "(.{1," + e + "})( +|$\n?)|(.{1," + e + "})";
          return t.match(RegExp(i, "g")).join("\n");
        }),
        (fi.prototype.getPrivateKey = function () {
          var t = "-----BEGIN RSA PRIVATE KEY-----\n";
          return ((t += this.wordwrap(this.getPrivateBaseKeyB64()) + "\n"), (t += "-----END RSA PRIVATE KEY-----"));
        }),
        (fi.prototype.getPublicKey = function () {
          var t = "-----BEGIN PUBLIC KEY-----\n";
          return ((t += this.wordwrap(this.getPublicBaseKeyB64()) + "\n"), (t += "-----END PUBLIC KEY-----"));
        }),
        (fi.prototype.hasPublicKeyProperty = function (t) {
          return ((t = t || {}), t.hasOwnProperty("n") && t.hasOwnProperty("e"));
        }),
        (fi.prototype.hasPrivateKeyProperty = function (t) {
          return (
            (t = t || {}),
            t.hasOwnProperty("n") &&
              t.hasOwnProperty("e") &&
              t.hasOwnProperty("d") &&
              t.hasOwnProperty("p") &&
              t.hasOwnProperty("q") &&
              t.hasOwnProperty("dmp1") &&
              t.hasOwnProperty("dmq1") &&
              t.hasOwnProperty("coeff")
          );
        }),
        (fi.prototype.parsePropertiesFrom = function (t) {
          ((this.n = t.n),
            (this.e = t.e),
            t.hasOwnProperty("d") &&
              ((this.d = t.d),
              (this.p = t.p),
              (this.q = t.q),
              (this.dmp1 = t.dmp1),
              (this.dmq1 = t.dmq1),
              (this.coeff = t.coeff)));
        }));
      var zi = function (t) {
        (fi.call(this),
          t &&
            ("string" == typeof t
              ? this.parseKey(t)
              : (this.hasPrivateKeyProperty(t) || this.hasPublicKeyProperty(t)) && this.parsePropertiesFrom(t)));
      };
      ((zi.prototype = new fi()), (zi.prototype.constructor = zi));
      var Zi = function (t) {
        ((t = t || {}),
          (this.default_key_size = parseInt(t.default_key_size) || 1024),
          (this.default_public_exponent = t.default_public_exponent || "010001"),
          (this.log = t.log || !1),
          (this.key = null));
      };
      ((Zi.prototype.setKey = function (t) {
        (this.log && this.key && console.warn("A key was already set, overriding existing."), (this.key = new zi(t)));
      }),
        (Zi.prototype.setPrivateKey = function (t) {
          this.setKey(t);
        }),
        (Zi.prototype.setPublicKey = function (t) {
          this.setKey(t);
        }),
        (Zi.prototype.decrypt = function (t) {
          try {
            return this.getKey().decrypt(Ri(t));
          } catch (e) {
            return !1;
          }
        }),
        (Zi.prototype.encrypt = function (t) {
          try {
            return Si(this.getKey().encrypt(t));
          } catch (e) {
            return !1;
          }
        }),
        (Zi.prototype.getKey = function (t) {
          if (!this.key) {
            if (((this.key = new zi()), t && "[object Function]" === {}.toString.call(t)))
              return void this.key.generateAsync(this.default_key_size, this.default_public_exponent, t);
            this.key.generate(this.default_key_size, this.default_public_exponent);
          }
          return this.key;
        }),
        (Zi.prototype.getPrivateKey = function () {
          return this.getKey().getPrivateKey();
        }),
        (Zi.prototype.getPrivateKeyB64 = function () {
          return this.getKey().getPrivateBaseKeyB64();
        }),
        (Zi.prototype.getPublicKey = function () {
          return this.getKey().getPublicKey();
        }),
        (Zi.prototype.getPublicKeyB64 = function () {
          return this.getKey().getPublicBaseKeyB64();
        }),
        (t.JSEncrypt = Zi));
    })(passport.lib.RSAExport),
    (passport.lib.RSA = passport.lib.RSAExport.JSEncrypt));
  var passport = passport || window.passport || {};
  ((passport.err = passport.err || {}),
    (function (e) {
      var i = null,
        s = "",
        r = new RegExp("(^|&)lang=([^&]*)(&|$)", "i"),
        o = window.location.search.substr(1).match(r);
      (null != o && (s = unescape(o[2])),
        (i = "function" === (typeof e.getCurrent).toLowerCase() ? e.getCurrent() : { errMsg: {}, labelText: {} }));
      var a = "https://ufosdk.baidu.com/bailingPC/getEntryPath/eG8Bo9LKhwNaa+UzF4RrYU-3KF8SAV35kX+Zqjzyp8U=",
        n = {
          "zh-CN": {
            400604: { msg: "请勿使用物联网卡手机号", field: "phone" },
            "-1": {
              msg: '系统错误,请您稍后再试,<a href="http://passport.baidu.com/v2/?ucenterfeedback#{urldata}#login"  target="_blank">帮助中心</a>',
              field: "",
            },
            1: { msg: "您输入的账号格式不正确", field: "userName" },
            2: {
              msg: '用户名或密码有误，请重新输入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密码</a>',
              field: "userName",
            },
            3: { msg: "验证码不存在或已过期,请重新输入", field: "" },
            4: {
              msg: "账号或密码错误，请重新输入或者<a href='http://passport.baidu.com/?getpassindex#{urldata}'  target='_blank' >找回密码</a>",
              field: "password",
            },
            5: { msg: "", field: "" },
            6: { msg: "您输入的验证码有误", field: "verifyCode" },
            7: {
              msg: "用户名或密码有误，请重新输入或<a href='http://passport.baidu.com/?getpassindex#{urldata}'  target='_blank' >找回密码</a>",
              field: "password",
            },
            16: {
              msg: '您的账号因安全问题已被限制登录,<a href="http://passport.baidu.com/v2/?ucenterfeedback#{urldata}#login"  target="_blank" >帮助中心</a>',
              field: "",
            },
            257: { msg: "请输入验证码", field: "verifyCode" },
            100027: { msg: "百度正在进行系统升级，暂时不能提供服务，敬请谅解", field: "" },
            120016: { msg: "", field: "" },
            18: { msg: "", field: "" },
            19: { msg: "", field: "" },
            20: { msg: "", field: "" },
            21: { msg: "没有登录权限", field: "" },
            22: { msg: "", field: "" },
            23: { msg: "", field: "" },
            24: { msg: "百度正在进行系统升级，暂时不能提供服务，敬请谅解", field: "" },
            25: { msg: "", field: "" },
            400031: { msg: "请在弹出的窗口操作,或重新登录", field: "" },
            400032: { msg: "", field: "" },
            400034: { msg: "", field: "" },
            401007: { msg: "您的手机号关联了其他账号，请选择登录", field: "" },
            120021: { msg: "登录失败,请在弹出的窗口操作,或重新登录", field: "" },
            500010: { msg: "登录过于频繁,请24小时后再试", field: "" },
            200010: { msg: "验证码不存在或已过期", field: "" },
            100005: { msg: "系统错误,请您稍后再试", field: "" },
            120019: { msg: "请在弹出的窗口操作,或重新登录", field: "userName" },
            110024: { msg: "此账号暂未激活,<a href='#{gotourl}' >重发验证邮件</a>", field: "" },
            100023: { msg: "系统升级中，为了您的账号安全，您可以使用二维码登录或者更换其他浏览器继续登录", field: "" },
            17: {
              msg: '您的账号已锁定,请<a href="http://passport.baidu.com/v2/?ucenterfeedback#login_10" target="_blank">解锁</a>后登录',
              field: "userName",
            },
            400401: { msg: "", field: "" },
            400037: { msg: "", field: "" },
            50023: { msg: "1个手机号30日内最多换绑3个账号", field: "" },
            50024: { msg: "注册过于频繁，请稍候再试", field: "" },
            50025: { msg: "注册过于频繁，请稍候再试；也可以通过上行短信的方式进行注册", field: "" },
            50028: {
              msg: '账号或密码多次输错，请3个小时之后再试或<a href="http://passport.baidu.com/?getpassindex&getpassType=financePwdError#{urldata}"  target="_blank">找回密码</a>',
              field: "",
            },
            50029: {
              msg: '账号或密码多次输错，请3个小时之后再试或<a href="http://passport.baidu.com/?getpassindex&getpassType=pwdError#{urldata}"  target="_blank">找回密码</a>',
              field: "",
            },
            400714: {
              msg: '系统检测你的账号疑似被盗，存在安全风险<a href="http://passport.baidu.com/?getpassindex&getpassType=pwdError#{urldata}"  target="_blank">请重置密码</a>',
              field: "",
            },
            50030: { msg: "抱歉，该手机号的申请次数已达当日上限，请更换手机号", field: "" },
            50031: { msg: "抱歉，该手机号的申请次数已达当月上限，请更换手机号", field: "" },
            50032: { msg: "抱歉，该手机号的申请次数已达本季度上限，请更换手机号", field: "" },
            50035: { msg: "系统错误,请您稍后再试(50035)", field: "" },
            400413: { msg: "", field: "" },
            400414: { msg: "", field: "" },
            400415: {
              msg: "账号存在风险，为了您的账号安全，请到百度钱包/理财/地图任一APP登录并完成验证，谢谢",
              field: "",
            },
            400500: { msg: "您登录的账号已注销，请登录其他账号或重新注册", field: "" },
            72200: { msg: "您的账号因安全风险被冻结，请前往APP登录页面底部帮助-账号解冻操作", field: "" },
            96001: { msg: "您的账号因违反百度用户协议被限制登录", field: "" },
            100060: { msg: "因系统升级暂不支持注册，预计6月7日恢复服务", field: "" },
            400702: { msg: "该账号已关闭手机号登录，请使用账号密码登录", field: "" },
            400704: { msg: "", field: "" },
            400706: { msg: "", field: "" },
            50043: { msg: "当前浏览器版本过低，请升级或更换浏览器后重试", field: "" },
            923002: { msg: "当前身份证无法识别", field: "" },
            97101: { msg: "系统繁忙，请稍候再试", field: "" },
            400703: { msg: "手机号已被运营商二次放号，请前往https://passport.baidu.com/v2/?reg注册新账号", field: "" },
            50045: { msg: "您的手机因安全问题已被限制" },
            50046: { msg: "您的ip因安全问题已被限制" },
            50044: { msg: "您的网络环境存在安全风险，请切换网络重试" },
            50047: { msg: "暂不支持海外及港澳台地区注册" },
            50048: { msg: "您当前的操作存在安全风险，请稍后再试" },
            26: { msg: "暂不支持海外及港澳台地区注册" },
            100073: { msg: "网盘账号系统正在升级，现暂停新账号注册，将于2月28日恢复开放，对此给您带来不便深表歉意！" },
            400012: { msg: '您的账号因违反百度用户协议被限制登录<a href="' + a + '"> 去申诉</a>' },
            400044: { msg: '您的账号因违反百度用户协议被限制登录<a href="' + a + '"> 去申诉</a>' },
            402012: {
              msg: "您百度账号存在安全风险，已被冻结，请到移动端app百度账号登录页面底部帮助-账号解冻按照页面提示操作解冻",
            },
            400608: { msg: "该账号无法在国外登录" },
            50000: { msg: "存在安全风险，请稍后再试" },
            50093: { msg: "系统正在升级维护，请您稍后重试", field: "" },
          },
          zh_HK: {
            "-1": {
              msg: '系統錯誤,請您稍後再試,<a href="http://passport.baidu.com/v2/?ucenterfeedback#{urldata}#login"  target="_blank">幫助中心</a>',
              field: "",
            },
            1: { msg: "您輸入的帳號格式不正確", field: "userName" },
            2: {
              msg: '用戶名或密碼有誤，請重新輸入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密碼</a>',
              field: "userName",
            },
            3: { msg: "驗證碼不存在或已過期,請重新輸入", field: "" },
            4: {
              msg: "帳號或密碼錯誤，請重新輸入或者<a href='http://passport.baidu.com/?getpassindex#{urldata}'  target='_blank' >找回密碼</a>",
              field: "password",
            },
            5: { msg: "", field: "" },
            6: { msg: "您輸入的驗證碼有誤", field: "verifyCode" },
            7: {
              msg: "用戶名或密碼有誤，請重新輸入或<a href='http://passport.baidu.com/?getpassindex#{urldata}'  target='_blank' >找回密碼</a>",
              field: "password",
            },
            16: {
              msg: '您的帳號因安全問題已被限制登錄,<a href="http://passport.baidu.com/v2/?ucenterfeedback#{urldata}#login"  target="_blank" >幫助中心</a>',
              field: "",
            },
            257: { msg: "請輸入驗證碼", field: "verifyCode" },
            100027: { msg: "百度正在進行系統升級，暫時不能提供服務，敬請諒解", field: "" },
            120016: { msg: "", field: "" },
            18: { msg: "", field: "" },
            19: { msg: "", field: "" },
            20: { msg: "", field: "" },
            21: { msg: "沒有登錄權限", field: "" },
            22: { msg: "", field: "" },
            23: { msg: "", field: "" },
            24: { msg: "百度正在進行系統升級，暫時不能提供服務，敬請諒解", field: "" },
            25: { msg: "", field: "" },
            400031: { msg: "請在彈出的窗口操作,或重新登錄", field: "" },
            400032: { msg: "", field: "" },
            400034: { msg: "", field: "" },
            401007: { msg: "您的手機號關聯了其他帳號，請選擇登錄", field: "" },
            120021: { msg: "登錄失敗,請在彈出的窗口操作,或重新登錄", field: "" },
            500010: { msg: "登錄過於頻繁,請24小時後再試", field: "" },
            200010: { msg: "驗證碼不存在或已過期", field: "" },
            100005: { msg: "系統錯誤,請您稍後再試", field: "" },
            120019: { msg: "請在彈出的窗口操作,或重新登錄", field: "userName" },
            110024: { msg: "此帳號暫未激活,<a href='#{gotourl}' >重發驗證郵件</a>", field: "" },
            100023: { msg: "系統升級中，為了您的帳號安全，您可以使用二維碼登錄或者更換其他瀏覽器繼續登錄", field: "" },
            17: {
              msg: '您的帳號已鎖定,請<a href="http://passport.baidu.com/v2/?ucenterfeedback#login_10" target="_blank">解鎖</a>後登錄',
              field: "userName",
            },
            400401: { msg: "", field: "" },
            400037: { msg: "", field: "" },
            50023: { msg: "1個手機號30日內最多換綁3個賬號", field: "" },
            50024: { msg: "註冊過於頻繁，請稍候再試", field: "" },
            50025: { msg: "註冊過於頻繁，請稍候再試；也可以通過上行短信的方式進行註冊", field: "" },
            50028: {
              msg: '帳號或密碼多次輸錯，請3個小時之後再試或<a href="http://passport.baidu.com/?getpassindex&getpassType=financePwdError#{urldata}"  target="_blank">找回密碼</a>',
              field: "",
            },
            50029: {
              msg: '帳號或密碼多次輸錯，請3個小時之後再試或<a href="http://passport.baidu.com/?getpassindex&getpassType=pwdError#{urldata}"  target="_blank">找回密碼</a>',
              field: "",
            },
            400714: {
              msg: '系統檢測你的帳號疑似被盜，存在安全風險<a href="http://passport.baidu.com/?getpassindex&getpassType=pwdError#{urldata}"  target="_blank">請重置密碼</a>',
              field: "",
            },
            50030: { msg: "抱歉，該手機號的申請次數已達當日上限，請更換手機號", field: "" },
            50031: { msg: "抱歉，該手機號的申請次數已達當月上限，請更換手機號", field: "" },
            50032: { msg: "抱歉，該手機號的申請次數已達本季度上限，請更換手機號", field: "" },
            50035: { msg: "系統錯誤,請您稍後再試(50035)", field: "" },
            400413: { msg: "", field: "" },
            400414: { msg: "", field: "" },
            400415: {
              msg: "帳號存在風險，為了您的帳號安全，請到百度錢包/理財/地圖任一APP登錄並完成驗證，謝謝",
              field: "",
            },
            400500: { msg: "您登錄的帳號已註銷，請登錄其他帳號或重新註冊", field: "" },
            72200: { msg: "您的賬號因安全風險被凍結，請前往APP登錄頁面底部幫助-賬號解凍操作", field: "" },
            96001: { msg: "您的帳號因違反百度用戶協議被限制登錄", field: "" },
            100060: { msg: "因系統升級暫不支持註冊，預計6月7日恢復服務", field: "" },
            400702: { msg: "該帳號已關閉手機號登錄，請使用帳號密碼登錄", field: "" },
            400704: { msg: "", field: "" },
            400706: { msg: "", field: "" },
            50043: { msg: "當前瀏覽器版本過低，請升級或更換瀏覽器後重試", field: "" },
            923002: { msg: "當前身份證無法識別", field: "" },
            97101: { msg: "系統繁忙，請稍候再試", field: "" },
            400703: { msg: "手機號已被運營商二次放號，請前往https://passport.baidu.com/v2/?reg註冊新帳號", field: "" },
            50045: { msg: "您的手機因安全問題已被限制" },
            50046: { msg: "您的ip因安全問題已被限制" },
            50044: { msg: "您的網絡環境存在安全風險，請切換網絡重試" },
            50047: { msg: "暫不支持海外及港澳臺地區註冊" },
            50048: { msg: "您當前的操作存在安全風險，請稍後再試" },
            26: { msg: "暫不支持海外及港澳臺地區註冊" },
            100073: { msg: "網盤帳號系統正在升級，現暫停新帳號註冊，將於2月28日恢復開放，對此給您帶來不便深表歉意！" },
            400012: { msg: '您的帳號因違反百度用戶協議被限制登錄<a href="' + a + '"> 去申訴</a>' },
            400044: { msg: '您的帳號因違反百度用戶協議被限制登錄<a href="' + a + '"> 去申訴</a>' },
            402012: {
              msg: "您百度賬號存在安全風險，已被凍結，請到移動端app百度賬號登錄頁面底部幫助-賬號解凍按照頁面提示操作解凍",
            },
            400608: { msg: "该账号无法在国外登录" },
            50000: { msg: "存在安全风险，请稍后再试" },
            50093: { msg: "系統正在升級維護，請您稍後重試", field: "" },
          },
          en: {
            400604: { msg: "Do not use IoT phone number", field: "phone" },
            "-1": { msg: "System error", field: "" },
            1: { msg: "The account format is incorrect", field: "userName" },
            2: { msg: "The username or password is incorrect", field: "userName" },
            3: { msg: "The code not exist or has expired", field: "" },
            4: { msg: "Wrong account or password", field: "password" },
            5: { msg: "", field: "" },
            6: { msg: "The verify code is incorrect", field: "verifyCode" },
            7: { msg: "Incorrect username or password", field: "password" },
            16: { msg: "The account has been restricted for security reasons", field: "" },
            257: { msg: "Please enter the verify code", field: "verifyCode" },
            100027: {
              msg: "The system upgrading and the service is temporarily unavailable. Please understand",
              field: "",
            },
            120016: { msg: "", field: "" },
            18: { msg: "", field: "" },
            19: { msg: "", field: "" },
            20: { msg: "", field: "" },
            21: { msg: "No login permission", field: "" },
            22: { msg: "", field: "" },
            23: { msg: "", field: "" },
            24: {
              msg: "The system is upgrading and the service is temporarily unavailable. Please understand",
              field: "",
            },
            25: { msg: "", field: "" },
            400031: { msg: "log in again", field: "" },
            400032: { msg: "", field: "" },
            400034: { msg: "", field: "" },
            401007: { msg: "The mobile is associated with another account, please choose to log in", field: "" },
            120021: { msg: "Login failed, please operate in the window or log in again", field: "" },
            500010: { msg: "Frequent, try after 24 hours", field: "" },
            200010: { msg: "The code does not exist or has expired", field: "" },
            100005: { msg: "System error, try again later", field: "" },
            120019: { msg: "please operate in the window or log in again", field: "userName" },
            110024: { msg: "The account is not activated,<a href='#{gotourl}' >Resend email</a>", field: "" },
            100023: {
              msg: "The system is being upgraded. For the safety of your account, you can switch to another browser to continue logging in.",
              field: "",
            },
            17: { msg: "Your account has been locked", field: "userName" },
            400401: { msg: "", field: "" },
            400037: { msg: "", field: "" },
            50023: { msg: "A phone number can be bound to up to 3 accounts within 30 days", field: "" },
            50024: { msg: "Registration is too frequent, please try again later", field: "" },
            50025: { msg: "Registration is too frequent, please try again later", field: "" },
            50028: { msg: "Multiple errors, please try again after 3 hours", field: "" },
            50029: { msg: "Multiple errors, please try again after 3 hours", field: "" },
            400714: { msg: "there is a security risk", field: "" },
            50030: {
              msg: "The number of applications has reached the daily limit, please change your phone number",
              field: "",
            },
            50031: {
              msg: "The number of applications has reached the daily limit, please change your phone number",
              field: "",
            },
            50032: {
              msg: "The number of applications has reached the daily limit, please change your phone number",
              field: "",
            },
            50035: { msg: "System error, please try again later (50035)", field: "" },
            400413: { msg: "", field: "" },
            400414: { msg: "", field: "" },
            400415: { msg: "Account is at risk", field: "" },
            400500: { msg: "The account has been cancelled. Please log in another or register again.", field: "" },
            72200: { msg: "The account cannot be logged in temporarily", field: "" },
            96001: { msg: "Your account has been restricted due to violation of Baidu User Agreement", field: "" },
            100060: { msg: "Registration is not supported due to system upgrade", field: "" },
            400702: { msg: "The account has closed phone login, please log in with your account password", field: "" },
            400704: { msg: "", field: "" },
            400706: { msg: "", field: "" },
            50043: {
              msg: "The current browser version is too low, please upgrade or change the browser and try again",
              field: "",
            },
            923002: { msg: "The current ID card cannot be recognized", field: "" },
            97101: { msg: "System error, please try again later", field: "" },
            400703: { msg: "The mobile phone number has been released by the operator for the second time", field: "" },
            50045: { msg: "Your phone has been restricted due to security issues" },
            50046: { msg: "Your IP has been restricted due to security issues" },
            50044: { msg: "The network ha s risks. Please switch networks and try again." },
            50047: { msg: "Registration in overseas, Hong Kong, Macau and Taiwan is not supported yet" },
            50048: { msg: "The operation has risks, please try again later" },
            26: { msg: "Registration in overseas, Hong Kong, Macau and Taiwan is not supported yet" },
            100073: { msg: "The system is being upgraded, and account registration is suspended" },
            400012: { msg: "Your account has been restricted due to violation of the user agreement." },
            400044: { msg: "Your account has been restricted due to violation of the user agreement." },
            402012: { msg: "There is a security risk in your account" },
            400608: { msg: "The account cannot be logged in overseas" },
            50000: { msg: "There is a security risk. Please try again later" },
            50093: { msg: "System is being upgraded, please try again later", field: "" },
          },
        },
        t = {
          "zh-CN": {
            500002: { msg: "您输入的验证码有误", field: "verifyCode" },
            500018: { msg: "验证码已失效，请重试", field: "verifyCode" },
          },
          zh_HK: {
            500002: { msg: "您輸入的驗證碼有誤", field: "verifyCode" },
            500018: { msg: "驗證碼已失效，請重試", field: "verifyCode" },
          },
          en: {
            500002: { msg: "The code is incorrect", field: "verifyCode" },
            500018: { msg: "The code has expired, please try again", field: "verifyCode" },
          },
        },
        d = {
          "zh-CN": {
            agree: "阅读并接受",
            baiduUserProtocal: "百度用户协议",
            userProtocal: "用户协议",
            baiduUserPersonal: "隐私政策",
            childAgreement: "儿童个人信息保护声明",
            baiduUsePassport: "百度账号使用说明",
            verifyCode: "验证码",
            verifyCodeStaErr: "您输入的验证码有误",
            verifyCodeLenErr: "您输入的验证码有误",
            unReceiveSmsCode: "没收到短信?",
            captcha: "验证码",
            captchaAlt: "验证码图片",
            captchaChange: "换一张",
            memberPassLabel: "下次自动登录",
            login: "登录",
            fgtPwd: "忘记密码？",
            feedback: "帮助中心",
            register: "立即注册",
            phoneNum: "手机号",
            account: "邮箱",
            userName: "手机号/用户名/邮箱",
            enterUserName: "请您输入用户名",
            password: "密码",
            passwordResetWarnNo:
              '用户名或密码有误，请重新输入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密码</a>',
            passwordResetSms: '<a id="passwordResetSms">短信登录</a>,或者',
            passwordResetWarn:
              '用户名或密码有误，请重新输入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密码</a>',
            passwordResetIn: "个月以内",
            passwordResetOut: "个月以前",
            unameMailLengthError: "邮箱过长,请重新输入",
            unameInputError: "邮箱格式错误,若未绑定邮箱,请使用用户名登录",
            smsPhone: "手机号",
            smsPhoneMsg: "请输入手机号",
            smsVerifyPlaceholder: "验证码",
            smsVerifyCode: "发送验证码",
            toSmsBtn: "短信快捷登录",
            logining: "登录中...",
            loginsuccess: "登录成功",
            submitTimeup: "登录超时,请稍后再试",
            backToLogin: "用户名密码登录",
            footerBackToLogin: "用户名登录",
            qrcodeTitle: "请使用",
            qrcodeHref:
              '<a class="pass-link" href="https://mo.baidu.com/wuxian/?from=1019447z" target="new">百度App</a>',
            qrcodeText: "扫码登录",
            qrcodeMsg: "百度技术加密，保障您的隐私安全",
            qrcodeListaq: "安全",
            qrcodeListgx: "高效",
            qrcodeListbj: "便捷",
            appName: "百度App",
            appHref: "http://xbox.m.baidu.com/mo/",
            sysError: "系统错误，休息一会儿，请稍后再试",
            sysUpdate: "服务正在升级中,请您稍后再试",
            cookieDisable:
              "开启Cookie之后才能登录,<a href='http://passport.baidu.com/v2/?ucenterfeedback#login'  target='_blank' >如何开启</a>?",
            captchaErr: "动态密码错误",
            confirmVerCodeEmpty: "验证码为空",
            foreignToLogin: "账号密码登录",
            foreignMobileError: "手机号码格式不正确",
            internetOfThingsMobileError: "请勿使用物联网手机号",
            foreignMobileMsg: "<span>请选择您的国家地区</span>",
            foreignMobileLink: "海外手机号",
            phoenixBtn: "第三方登录",
            qrcodeBtn: "扫码登录",
            QrcodeSuccessTip: "扫描成功",
            QrcodeSuccessMsg: "请在手机端确认登录",
            QrcodeErrorTip: "网络连接失败",
            QrcodeErrorMsg: "请稍候再试",
            QrcodeRiskTip: "操作频繁",
            QrcodeRefreshTip: "二维码已失效",
            QrcodeRefreshBtn: "点击刷新",
            QrcodeLoadTip: "二维码加载失败",
            nopassLead:
              '该账号尚未设置密码，请先<a href="http://passport.baidu.com/?getpassindex" target="_blank" >设置密码</a>后在登录',
            resend: "重新发送",
            userNameLogin: "用户名登录",
            emailLogin: "邮箱登录",
            forgetEmail: "忘记邮箱?使用",
            fillPhonePls: "请填写手机号",
            agreeCheckError: "请勾选下方协议",
            thirdCheckDialogTitle: "提示",
            thirdCheckDialogBody: "为了更好的使用服务，登录前请阅读并同意以下协议",
            thirdCheckDialogAgree: "同意",
            thirdCheckDialogDisagree: "不同意",
            userNameTitle: "用户名",
            smsLogin: "短信登录",
            fillVerifyCodePls: "请填写验证码",
            forgetUserName: "忘记用户名?使用",
            verifyEqualLogin: "验证即登录，未注册将自动创建百度账号",
            cantReceiveSms: "收不到验证码",
            checkIfBlocked: "1.请检查短信是否被安全软件拦截",
            overTimeArrive: "2.运营商网络原因，短信可能延迟到达",
            reguest: "建议您",
            switchLoginMethod: "切换登录方式",
            usePhoneSendSms: "使用该手机号，编辑短信",
            shouldLogSuccess: "即可登录成功",
            introductSms: "（短信费用请参考运营商资费标准）",
            rule1: "建议使用当前手机号，编辑短信 8-14 位字符（支持数字/字母/符号），",
            rule2: "作为登录密码，发送至" + window.upsmsStore.reg_upsms + "，即可注册成功",
            mainland: "中国境内（不含港澳台）",
            clientOnekeytip1: "您已登录",
            clientOnekeytip2: "客户端，点击直接登录",
            clientOnekeySumbit: "直接登录",
            clientOnekeyChangeLogin: "账号登录",
            networkError: "网络连接失败，请检查您的网络设置并稍后再试。",
            qrcodeCancelTip: "已取消授权",
            useWxMiniLogin: "请使用微信扫码登录",
            click: "点击",
            changeBaiduQrCode: "切换百度APP扫码登录",
            changeWxQrCode: "切换微信APP扫码登录",
            wxQrcodeSuccessTip: "微信扫码成功",
            wxMiniSureLogin: "请在小程序端确认登录",
            and: "和",
            voiceCode: "获取语音验证码",
            unblocking: "去解禁",
            or: "或",
            sendTo: "发送至",
            vCode: "语音验证码",
            Sent: "已发送",
            unbindingSmsTip: "物联网卡手机号未修改，被限制登录",
            unbindingSmsTip1: "您账号存在安全风险被限制登录",
            dearUser: "亲爱的用户，",
            securityTips: "安全提示",
            choiceuserTip: "为了确保您的账号安全，请先确认您输入的账号是用户名还是手机号",
            upsmsTip: "您当前的操作存在风险，请用该手机号",
            confirmText: "确定",
            continueToValidate: "继续验证",
            safety: "安全验证",
            fillPicture: "请填写图中的验证码",
            youCan: "您可以",
            youAlsoCan: "您也可以",
            jump: "跳过此步骤",
            bindPhone: "绑定手机",
            bindPassMail: "绑定密保邮箱",
            confirmWidgetTip1: "近日，某邮箱账户存在被破解的可能性，建议您使用其他",
            confirmWidgetTip2: "邮箱绑定百度账号，或及时修改您的百度账号登陆密码并",
            confirmWidgetTip3: "绑定手机，以确保账户安全。",
            enterBind: "请输入您要绑定的手机号码",
            setBindPhone: "建议设置为绑定手机",
            checkUnVPhone: "检测到您有待验证手机",
            verifySuccess: "验证成功",
            bindTip: "为了您的账号安全，建议设置为绑定手机号",
            warnMsg1: "您的账号密码输入错误次数达到上限，为保障账号安全，登录前需验证身份。",
            warnMsg2: "您的网络环境存在安全风险，为保障账号安全，登录前需验证身份。",
            warnMsg3: "您的账号长时间未登录，为保障账号安全，登录前需验证身份。",
            warnMsg4: "您的账号可能存在安全隐患，为保障您的账号安全，登录前需验证身份。",
            warnMsg5: "您的账号因批量或者使用非法软件注册被冻结，登录前需验证身份。",
            warnMsg6: "您操作频度过于频繁，为保障账号安全，登录前需验证身份。",
            warnMsg7: "您本次的登录地存在异常，为保障本次操作安全，登录前需验证身份。",
            warnMsg8: "您的账号存在安全风险，为保障账号安全，登录前需验证身份。",
            loginProtect: "登录保护",
            loginTip1: "您已开启登录保护功能，为保障账号安全，登录前需验证身份。",
            loginSafeV: "登录安全验证",
            loginTip2: "存在安全风险，请稍后再试",
            findPassword: "找回密码",
            reLogin: "请重新登录",
            loginUnfreeze: "登录解冻验证",
            loginTip3: "您的账号因密码输入错误次数过多，为保障账号安全，登陆前需验证身份。",
            loginLimit: "登录密码错误已达上限",
            thressThour: "3小时后再试",
            registerTip: "您的手机号码尚未注册或未开启手机号登录，点击注册，帮您注册新的百度账号",
            deleteText: "删除该记录",
            regist: "注册",
            registTip: "手机号未注册或已被运营商二次放号，请重新注册新账号。",
            lock: "大小写锁定已打开",
            switchOverseasLogin: "切换谷歌登录等方式",
          },
          zh_HK: {
            agree: "閱讀並接受",
            baiduUserProtocal: "百度用戶協議",
            userProtocal: "用戶協議",
            baiduUserPersonal: "隱私政策",
            baiduUsePassport: "百度帳號使用說明",
            childAgreement: "兒童個人信息保護聲明",
            verifyCode: "驗證碼",
            verifyCodeStaErr: "您輸入的驗證碼有誤",
            verifyCodeLenErr: "您輸入的驗證碼有誤",
            unReceiveSmsCode: "没收到短信?",
            captcha: "驗證碼",
            captchaAlt: "驗證碼圖片",
            captchaChange: "換一張",
            memberPassLabel: "下次自動登錄",
            login: "登錄",
            fgtPwd: "忘記密碼？",
            feedback: "幫助中心",
            register: "立即註冊",
            phoneNum: "手機號",
            account: "郵箱",
            userName: "手機/郵箱/用戶名",
            enterUserName: "請您輸入用戶名",
            password: "密碼",
            passwordResetWarnNo:
              '用戶名或密碼有誤，請重新輸入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密碼</a>',
            passwordResetSms: '<a id="passwordResetSms">短信登錄</a>,或者',
            passwordResetWarn:
              '用戶名或密碼有誤，請重新輸入或<a href="http://passport.baidu.com/?getpassindex#{urldata}"  target="_blank" >找回密碼</a>',
            passwordResetIn: "個月以內",
            passwordResetOut: "個月以前",
            unameMailLengthError: "郵箱過長,請重新輸入",
            unameInputError: "郵箱格式錯誤,若未綁定郵箱,請使用用戶名登錄",
            smsPhone: "手機號",
            smsPhoneMsg: "請輸入手機號",
            smsVerifyPlaceholder: "驗證碼",
            smsVerifyCode: "發送驗證碼",
            toSmsBtn: "短信快捷登錄",
            logining: "登錄中...",
            loginsuccess: "登錄成功",
            submitTimeup: "登錄超時,請稍後再試",
            backToLogin: "用戶名密碼登錄",
            footerBackToLogin: "用戶名登錄",
            qrcodeTitle: "請使用",
            qrcodeHref:
              '<a class="pass-link" href="https://mo.baidu.com/wuxian/?from=1019447z" target="new">百度App</a>',
            qrcodeText: "掃碼登錄",
            qrcodeMsg: "百度技術加密，保障您的隱私安全",
            qrcodeListaq: "安全",
            qrcodeListgx: "高效",
            qrcodeListbj: "便捷",
            appName: "百度App",
            appHref: "http://xbox.m.baidu.com/mo/",
            sysError: "系統錯誤，休息一會兒，請稍後再試",
            sysUpdate: "服務正在升級中,請您稍後再試",
            cookieDisable:
              "開啟Cookie之後才能登錄,<a href='http://passport.baidu.com/v2/?ucenterfeedback#login'  target='_blank' >如何開啟</a>?",
            captchaErr: "動態密碼錯誤",
            confirmVerCodeEmpty: "驗證碼為空",
            foreignToLogin: "帳號密碼登錄",
            foreignMobileError: "手機號碼格式不正確",
            internetOfThingsMobileError: "請勿使用物聯網手機號",
            foreignMobileMsg: "<span>請選擇您的國家地區</span>",
            foreignMobileLink: "海外手機號",
            phoenixBtn: "第三方登錄",
            qrcodeBtn: "掃碼登錄",
            QrcodeSuccessTip: "掃描成功",
            QrcodeSuccessMsg: "請在手機端確認登錄",
            QrcodeErrorTip: "網絡連接失敗",
            QrcodeErrorMsg: "請稍候再試",
            QrcodeRiskTip: "操作频繁",
            QrcodeRefreshTip: "二維碼已失效",
            QrcodeRefreshBtn: "點擊刷新",
            QrcodeLoadTip: "二維碼加載失敗",
            nopassLead:
              '該帳號尚未設置密碼，請先<a href="http://passport.baidu.com/?getpassindex#{urldata}" target="_blank" >設置密碼</a>後在登錄',
            resend: "重新發送",
            userNameLogin: "用戶名登錄",
            emailLogin: "郵箱登錄",
            forgetEmail: "忘記郵箱?使用",
            fillPhonePls: "請填寫手機號",
            agreeCheckError: "請勾選下方協議",
            thirdCheckDialogTitle: "提示",
            thirdCheckDialogBody: "為了更好的使用服務，登錄前請閱讀並同意以下協議",
            thirdCheckDialogAgree: "同意",
            thirdCheckDialogDisagree: "不同意",
            userNameTitle: "用戶名",
            smsLogin: "短信登錄",
            fillVerifyCodePls: "請填寫驗證碼",
            forgetUserName: "忘記用戶名?使用",
            verifyEqualLogin: "驗證即登錄，未註冊將自動創建百度帳號",
            cantReceiveSms: "收不到驗證碼",
            checkIfBlocked: "1.請檢查短信是否被安全軟件攔截",
            overTimeArrive: "2.運營商網絡原因，短信可能延遲到達",
            reguest: "建議您",
            switchLoginMethod: "切換登錄方式",
            usePhoneSendSms: "使用該手機號，編輯短信",
            shouldLogSuccess: "即可登錄成功",
            introductSms: "（短信費用請參考運營商資費標準）",
            rule1: "建議使用當前手機號，編輯短信 8-14 位字符（支持數字/字母/符號），",
            rule2: "作為登錄密碼，發送至" + window.upsmsStore.reg_upsms + "，即可註冊成功",
            mainland: "中國境內（不含港澳臺）",
            clientOnekeytip1: "您已登錄",
            clientOnekeytip2: "客戶端，點擊直接登錄",
            clientOnekeySumbit: "直接登錄",
            clientOnekeyChangeLogin: "賬號登錄",
            networkError: "網絡連接失敗，請檢查您的網絡設置並稍後再試。",
            qrcodeCancelTip: "已取消授權",
            useWxMiniLogin: "請使用微信掃碼登錄",
            click: "點擊",
            changeBaiduQrCode: "切換百度APP掃碼登錄",
            changeWxQrCode: "切換微信APP掃碼登錄",
            wxQrcodeSuccessTip: "微信掃碼成功",
            wxMiniSureLogin: "請在小程序端確認登錄",
            and: "和",
            voiceCode: "獲取語音驗證碼",
            unblocking: "去解禁",
            or: "或",
            sendTo: "發送至",
            vCode: "語音驗證碼",
            Sent: "已發送",
            unbindingSmsTip: "物聯網卡手機號未修改，被限制登錄",
            unbindingSmsTip1: "您賬號存在安全風險被限制登錄",
            dearUser: "親愛的用戶，",
            securityTips: "安全提示",
            choiceuserTip: "爲了確保您的賬號安全，請先確認您輸入的賬號是用戶名還是手機號",
            confirmText: "確定",
            continueToValidate: "繼續驗證",
            upsmsTip: "您當前的操作存在風險，請用該手機號",
            safety: "安全驗證",
            fillPicture: "請填寫圖中的驗證碼",
            youCan: "您可以",
            youAlsoCan: "您也可以",
            jump: "跳過此步驟",
            bindPhone: "綁定手機",
            bindPassMail: "綁定密保郵箱",
            confirmWidgetTip1: "近日，某郵箱賬戶存在被破解的可能性，建議您使用其他",
            confirmWidgetTip2: "郵箱綁定百度賬號，或及時修改您的百度賬號登陸密碼並",
            confirmWidgetTip3: "綁定手機，以確保賬戶安全。",
            enterBind: "請輸入您要綁定的手機號碼",
            setBindPhone: "建議設置爲綁定手機",
            checkUnVPhone: "檢測到您有待驗証手機",
            verifySuccess: "驗証成功",
            bindTip: "爲了您的賬號安全，建議設置爲綁定手機號",
            warnMsg1: "您的賬號密碼輸入錯誤次數達到上限，爲保障賬號安全，登錄前需驗証身份。",
            warnMsg2: "您的網絡環境存在安全風險，爲保障賬號安全，登錄前需驗証身份。",
            warnMsg3: "您的賬號長時間未登錄，爲保障賬號安全，登錄前需驗証身份。",
            warnMsg4: "您的賬號可能存在安全隱患，爲保障您的賬號安全，登錄前需驗証身份。",
            warnMsg5: "您的賬號因批量或者使用非法軟件注冊被凍結，登錄前需驗証身份。",
            warnMsg6: "您操作頻度過於頻繁，爲保障賬號安全，登錄前需驗証身份。",
            warnMsg7: "您本次的登錄地存在異常，爲保障本次操作安全，登錄前需驗証身份。",
            warnMsg8: "您的賬號存在安全風險，爲保障賬號安全，登錄前需驗証身份。",
            loginProtect: "登錄保護",
            loginTip1: "您已開啟登錄保護功能，爲保障賬號安全，登錄前需驗証身份。",
            loginSafeV: "登錄安全驗証",
            loginTip2: "存在安全風險，請稍後再試",
            findPassword: "找回密碼",
            reLogin: "請重新登錄",
            loginUnfreeze: "登錄解凍驗証",
            loginTip3: "您的賬號因密碼輸入錯誤次數過多，爲保障賬號安全，登陸前需驗証身份。",
            loginLimit: "登錄密碼錯誤已達上限",
            thressThour: "3小時後再試",
            registerTip: "您的手機號碼尚未注冊或未開啟手機號登錄，點擊注冊，幫您注冊新的百度賬號",
            deleteText: "刪除該記錄",
            regist: "注冊",
            registTip: "手機號未注冊或已被運營商二次放號，請重新注冊新賬號。",
            lock: "大小冩鎖定已打開。",
            switchOverseasLogin: "切換谷歌登錄等方式",
          },
          en: {
            agree: "Read and accept",
            baiduUserProtocal: "Terms of Policy",
            userProtocal: "Terms of Policy",
            baiduUserPersonal: "Privacy Policy",
            childAgreement: "Child Protection Statement",
            baiduUsePassport: "account instructions",
            verifyCode: "Verify code",
            verifyCodeStaErr: "Verify code error",
            verifyCodeLenErr: "Verify code error",
            unReceiveSmsCode: "Can't get the code",
            captcha: "captcha",
            captchaAlt: "Captcha Image",
            captchaChange: "Change it",
            memberPassLabel: "Automatic login next time",
            login: "Login",
            fgtPwd: "Forget password",
            feedback: "Help Center",
            register: "Register",
            phoneNum: "Phone number",
            account: "Email",
            userName: "Username",
            password: "Password",
            passwordResetWarnNo: "Wrong username or password",
            passwordResetSms: '<a id="passwordResetSms">Login</a>,or',
            passwordResetWarn: "Wrong username or password",
            passwordResetIn: "month within",
            passwordResetOut: "month ago",
            unameMailLengthError: "Email is too long, please re-enter",
            unameInputError: "Incorrect email format or unbound email",
            smsPhone: "Phone number",
            smsPhoneMsg: "Phone number",
            smsVerifyPlaceholder: "Verify code",
            smsVerifyCode: "send",
            toSmsBtn: "SMS Login",
            logining: "logging in...",
            loginsuccess: "Login Successful",
            submitTimeup: "Login timeout, please try again later",
            backToLogin: "User name and password login",
            footerBackToLogin: "User Name Login",
            qrcodeTitle: "Please use",
            qrcodeHref: "",
            qrcodeText: "Scan code to log in",
            qrcodeMsg: "technology encryption to protect your privacy",
            qrcodeListaq: "safety",
            qrcodeListgx: "efficiently",
            qrcodeListbj: "convenience",
            appName: "",
            appHref: "",
            sysError: "System error",
            sysUpdate: "Services being upgraded",
            cookieDisable: "Enable cookies to log in",
            captchaErr: "Dynamic password error",
            confirmVerCodeEmpty: "Verify code is empty",
            foreignToLogin: "Account Login",
            foreignMobileError: "Phone number format error",
            internetOfThingsMobileError: "Do not use the IoT phone number",
            foreignMobileMsg: "<span>Please select your country</span>",
            foreignMobileLink: "Overseas Phone",
            phoenixBtn: "Third Party Login",
            qrcodeBtn: "Scan code to login",
            QrcodeSuccessTip: "Scanning Successful",
            QrcodeSuccessMsg: "Please confirm your login on your phone",
            QrcodeErrorTip: "Network Connection Failure",
            QrcodeErrorMsg: "Please try again later.",
            QrcodeRiskTip: "High frequency of operation",
            QrcodeRefreshTip: "QR code is no longer valid",
            QrcodeRefreshBtn: "Refresh",
            QrcodeLoadTip: "QR code failed to load",
            nopassLead: "No password set before login",
            resend: "resend",
            userNameLogin: "User Name Login",
            emailLogin: "Email Login",
            forgetEmail: "Forgot email? Use",
            fillPhonePls: "Please fill in phone number",
            agreeCheckError: "Please check the agreement below",
            thirdCheckDialogTitle: "Notice",
            thirdCheckDialogBody: "Please read and agree to the following agreements before logging in",
            thirdCheckDialogAgree: "Agree",
            thirdCheckDialogDisagree: "Disagree",
            userNameTitle: "Username",
            smsLogin: "SMS Login",
            fillVerifyCodePls: "Please fill in verify code",
            forgetUserName: "Forgot username? Use",
            verifyEqualLogin: "Login after verify, it will be created auto if not register",
            cantReceiveSms: "Can't get the verify code",
            checkIfBlocked: "1.Please check if the SMS is blocked by the software",
            overTimeArrive: "2.SMS may be delayed due to operator reason",
            reguest: "recommend you",
            switchLoginMethod: "Switch Login Method",
            usePhoneSendSms: "Use phone number, edit text message",
            shouldLogSuccess: "You can log in successfully",
            introductSms: "（Please refer to the operator's tariff for SMS charges）",
            rule1:
              "Recommend to use phone number and edit the SMS with 8-14 characters (support numbers/letters/symbols)",
            rule2:
              "as the login password, send it to" + window.upsmsStore.reg_upsms + "，You can register successfully",
            mainland: "Within China (excluding Hong Kong, Macao and Taiwan)",
            clientOnekeytip1: "You are logged in",
            clientOnekeytip2: "Client, click to log in",
            clientOnekeySumbit: "Login",
            clientOnekeyChangeLogin: "Account Login",
            networkError: "Network connection failed",
            and: "and",
            voiceCode: "Get voice verify code",
            unblocking: "",
            or: "or",
            sendTo: "Send to",
            vCode: "voice verify code",
            Sent: "Sent",
            unbindingSmsTip: "IoT card phone number has not been changed and is restricted from logging in",
            unbindingSmsTip1: "Account has been restricted from log in due to risk",
            dearUser: "Dear User",
            securityTips: "Tips",
            choiceuserTip: "Please confirm whether you are entering username or phone number",
            confirmText: "Confirm",
            continueToValidate: "Continue verify",
            upsmsTip: "There is a risk, please use this phone number",
            safety: "safety certification",
            fillPicture: "Please fill in the verify code",
            youCan: "You can",
            jump: "Skip the step",
            bindPhone: " Bind your phone",
            youAlsoCan: "You can also",
            bindPassMail: "Bind email",
            confirmWidgetTip1: "Email are at risk,recommend that you use other",
            confirmWidgetTip2: "email bind acconut or change login password and",
            confirmWidgetTip3: "Bind phone to ensure safety",
            enterBind: "Please enter phone number you want to bind",
            setBindPhone: "Suggest set it as bind phone",
            checkUnVPhone: "Detect that your phone need to be verify",
            verifySuccess: "Verify success",
            bindTip: "For account security, it is recommended to set it to bind phone number",
            warnMsg1: "Password error limit reached, identity verify required",
            warnMsg2: "There is security risk and identity verify is required",
            warnMsg3: "Account has not been logged in for long time, identity verify is required",
            warnMsg4: "The account has security risks and requires identity verify",
            warnMsg5: "The account is frozen and identity verify is required",
            warnMsg6: "The operation is too frequent, identity verify is required",
            warnMsg7: "There is abnormal login location, identity verify is required",
            warnMsg8: "The account is at risk and identity verify is required",
            loginProtect: "Login protect",
            loginTip1: "Login protect enabled, identity verify required",
            loginSafeV: "Login verify",
            loginTip2: "There is security risk, please try again later",
            findPassword: "Retrieve password",
            reLogin: "Please log in again",
            loginUnfreeze: "Login unfreeze verify",
            loginTip3: "Password has been entered incorrectly too many times,identity verify is required",
            loginLimit: "Password error limit has been reached",
            thressThour: "Try again in 3 hours.",
            registerTip: "Your phone has not been registered or phone login is not enabled. To register a new account.",
            deleteText: "Delete the record",
            regist: "Register",
            registTip: "The phone is not registered, please register a new account.",
            lock: "Caps Lock is on",
            switchOverseasLogin: "Toggle Google login and more",
          },
        },
        g = {
          "zh-CN": {
            bindAction16: "您的账号存在安全风险，我们已经为您采取保护策略，建议您先绑定手机。",
            bindAction32: "快来绑定密保工具吧，提升账号安全性的同时可以快速找回密码。",
            bindAction34: "请绑定您的手机号码作为您的密保手机，提升账号安全性的同时还可以快速找回密码。",
            bindAction: "请绑定一个您的常用邮箱作为您的密保邮箱，提升账号安全性的同时还可以快速找回密码。",
            getWarnHtml: "系统检测到您的账号疑似被盗，存在安全风险。请尽快修改密码。",
          },
          zh_HK: {
            500002: { msg: "您輸入的驗證碼有誤", field: "verifyCode" },
            500018: { msg: "驗證碼已失效，請重試", field: "verifyCode" },
          },
          en: {
            bindAction16: "There are risks, it is recommended to bind phone first",
            bindAction32: "Bind the password protection tool to enhance security",
            bindAction34: "Bind your phone number for added security",
            bindAction: "Bind email for added security",
            getWarnHtml: "There is a risk, please change your password",
            500002: { msg: "You have entered the wrong code.", field: "verifyCode" },
            500018: { msg: "The validation code is invalid", field: "verifyCode" },
          },
        },
        l = {
          "zh-CN": {
            agree: "阅读并同意",
            user: "《百度用户协议》",
            personal: "《隐私政策》",
            passUse: "《百度账号使用说明》",
          },
          zh_HK: {
            agree: "閱讀並同意",
            user: "《百度用戶協議》",
            personal: "《隱私政策》",
            passUse: "《百度帳號使用說明》",
          },
          en: {
            agree: "Read and agree",
            user: "User Agreement",
            personal: "Privacy Policy",
            passUse: "Account instructions",
          },
        };
      e.getCurrent = function () {
        var e = "zh_HK" === s ? "zh_HK" : "en" === s ? "en" : "zh-CN",
          r = passport.passLang ? passport.passLang : e;
        return (
          i && i.errMsg && (i.errMsg.login = n[r]),
          i && i.errMsg && (i.errMsg.checkVerifycode = t[r]),
          i && i.labelText && (i.labelText.login = d[r]),
          (i.agreements = l[r]),
          (i.hooksText = g[r]),
          i
        );
      };
    })(passport.err));
  var passport = passport || window.passport || {};
  passport.data = passport.data || {};
  var passLang = passport && passport.passLang;
  !(function (e) {
    function t(e) {
      ((this._requests = []), (this._value = null), (this._exception = null), (this._isComplete = !1));
      var t = this;
      e(
        function (e) {
          t._fulfillPromise(e);
        },
        function (e) {
          t._breakPromise(e);
        },
      );
    }
    function n(e, t, n) {
      return t
        ? n
          ? function (n) {
              return (
                (n = n || {}),
                u.submit(l + t, r(n, e, d[e], g[e], !0), {
                  charset: "utf-8",
                  processData: function (t) {
                    if (t)
                      for (var n in t)
                        if (t.hasOwnProperty(n)) {
                          var r = t[n];
                          r && (t[n] = decodeURIComponent(r));
                        }
                    return a(e, t);
                  },
                })
              );
            }
          : function (n) {
              return u.jsonp(l + t, r(n, e, d[e], g[e], !1), {
                charset: "utf-8",
                processData: function (t) {
                  return a(e, t);
                },
              });
            }
        : s;
    }
    function r(e, t, n, r, a) {
      var i = a
          ? { staticpage: y.staticPage, charset: y.charset || document.characterSet || document.charset || "" }
          : {},
        o = h[t];
      if (o)
        for (var s in o) {
          if (o.hasOwnProperty(s)) {
            var u = o[s];
            i[s] = "function" == typeof u ? u(e) : u;
          }
          "verifypass" == s && (i[s] = decodeURIComponent(i[s]));
        }
      if (
        ((i.token = y.token || ""),
        (i.tpl = y.product || ""),
        (i.subpro = y.subpro || ""),
        (i.apiver = "v3"),
        (i.tt = new Date().getTime()),
        e)
      ) {
        ((n = n || {}), (r = r || {}));
        for (var s in e)
          if (e.hasOwnProperty(s)) {
            var l = r[s],
              p = l ? l(e[s], e) : e[s];
            ("string" == typeof p && (a && (p = decodeURIComponent(p)), m[s] || (p = c.trim(p))),
              (i[n[s] || s.toLowerCase()] = p));
          }
      }
      return i;
    }
    function a(t, n) {
      if ((e && e.traceID && e.traceID.getTraceID && e.traceID.getTraceID(n), n)) {
        var r = v[t];
        r && r(n);
        var a = n.errInfo,
          o = n,
          s = o;
        return (
          a
            ? (o.errInfo = i(t, a, o))
            : ((a = { no: n.err_no, msg: n.err_msg || "" }),
              delete o.err_no,
              delete o.err_msg,
              (s = { data: o, errInfo: i(t, a, o) })),
          s
        );
      }
      return n;
    }
    function i(e, t) {
      var n = O[w[e] || e];
      if (n && t && 0 != t.no) {
        var r = n[t.no] || n[-1];
        if (r) {
          var a = r.msg;
          ((t.msg = a), (t.field = r.field));
        }
      }
      return t;
    }
    function o(t) {
      if ((e && e.traceID && e.traceID.getTraceID && e.traceID.getTraceID(t), t)) {
        var n = t.errInfo,
          r = t;
        if (!n)
          for (var a in t)
            if (t.hasOwnProperty(a)) {
              var i = t[a];
              i && (t[a] = decodeURIComponent(i));
            }
        n ||
          ((n = { no: t.err_no, msg: t.err_msg || "" }),
          delete r.err_no,
          delete r.err_msg,
          (t = { data: r, errInfo: n }));
      }
      return t;
    }
    var s = function () {};
    ((t.prototype = {
      get_isComplete: function () {
        return this._isComplete;
      },
      get_value: function () {
        if (!this._isComplete) return void 0;
        if (this._exception) throw this._exception;
        return this._value;
      },
      call: function (e) {
        for (var t = [], n = 0, r = arguments.length - 1; r > n; n++) t[n] = arguments[n + 1];
        return this.when(function (n) {
          return n[e].apply(n, t);
        });
      },
      getValue: function (e) {
        return this.when(function (t) {
          return t[e];
        });
      },
      setValue: function (e, t) {
        this.whenOnly(function (n) {
          n[e] = t;
        });
      },
      when: function (e, n, r) {
        return t.when(this, e, n, r);
      },
      whenOnly: function (e, n, r) {
        t.whenOnly(this, e, n, r);
      },
      success: function (e, t) {
        return this.when(e, s, t);
      },
      fail: function (e, t) {
        return this.when(s, e, t);
      },
      _enqueueOne: function (e) {
        this._isComplete ? this._notify(e) : this._requests.push(e);
      },
      _notify: function (e) {
        this._exception
          ? e.breakPromise && e.breakPromise(this._exception)
          : e.fulfillPromise && e.fulfillPromise(this._value);
      },
      _notifyAll: function () {
        for (var e = 0, t = this._requests.length; t > e; e++) this._notify(this._requests[e]);
      },
      _fulfillPromise: function (e) {
        ((this._value = e), (this._exception = null), (this._isComplete = !0), this._notifyAll());
      },
      _breakPromise: function (e) {
        ((this._value = null),
          (this._exception = e || new Error("An error occured")),
          (this._isComplete = !0),
          this._notifyAll());
      },
    }),
      (t.when = function (e, n, r, a) {
        return new t(function (i, o) {
          t.make(e)._enqueueOne({
            fulfillPromise: function (e) {
              i(n ? n.call(a, e) : e);
            },
            breakPromise: function (e) {
              if (r)
                try {
                  i(r.call(a, e));
                } catch (t) {
                  o(t);
                }
              else o(e);
            },
          });
        });
      }),
      (t.whenOnly = function (e, n, r, a) {
        t.make(e)._enqueueOne({
          fulfillPromise: function (e) {
            n && n.call(a, e);
          },
          breakPromise: function (e) {
            r && r.call(a, e);
          },
        });
      }),
      (t.make = function (e) {
        return e instanceof t ? e : t.immediate(e);
      }),
      (t.immediate = function (e) {
        return new t(function (t) {
          t(e);
        });
      }));
    var c = {};
    (!(function (e) {
      var t = new RegExp("(^[\\s\\t\\xa0\\u3000]+)|([\\u3000\\xa0\\s\\t]+$)", "g");
      ((e.trim = function (e) {
        return String(e).replace(t, "");
      }),
        (e.getUniqueId = function (e) {
          return e + Math.floor(2147483648 * Math.random()).toString(36);
        }),
        (e.g = function (e) {
          return e
            ? "string" == typeof e || e instanceof String
              ? document.getElementById(e)
              : !e.nodeName || (1 != e.nodeType && 9 != e.nodeType)
                ? null
                : e
            : null;
        }),
        (e.getParent = function (t) {
          return ((t = e.g(t)), t.parentElement || t.parentNode || null);
        }),
        (e.encodeHTML = function (e) {
          return String(e)
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#39;");
        }),
        (e.array = e.array || {}),
        (e.array.indexOf = function (e, t, n) {
          var r = e.length;
          for (n = 0 | n, 0 > n && (n = Math.max(0, r + n)); r > n; n++) if (n in e && e[n] === t) return n;
          return -1;
        }),
        (e.browser = e.browser || {}),
        (e.browser.opera = /opera(\/| )(\d+(\.\d+)?)(.+?(version\/(\d+(\.\d+)?)))?/i.test(navigator.userAgent)
          ? +(RegExp.$6 || RegExp.$2)
          : void 0),
        (e.insertHTML = function (t, n, r) {
          t = e.g(t);
          var a, i;
          return (
            t.insertAdjacentHTML && !e.browser.opera
              ? t.insertAdjacentHTML(n, r)
              : ((a = t.ownerDocument.createRange()),
                (n = n.toUpperCase()),
                "AFTERBEGIN" == n || "BEFOREEND" == n
                  ? (a.selectNodeContents(t), a.collapse("AFTERBEGIN" == n))
                  : ((i = "BEFOREBEGIN" == n), a[i ? "setStartBefore" : "setEndAfter"](t), a.collapse(i)),
                a.insertNode(a.createContextualFragment(r))),
            t
          );
        }));
    })(c),
      (e.base = c));
    var u = {};
    !(function (n) {
      var r = "__bdpp_pstc__" + new Date().getTime(),
        a = r + "_form",
        i = r + "_ifr",
        o = function (e) {
          if ("object" == typeof e) {
            var t = [];
            for (var n in e) {
              var r = e[n];
              if (void 0 !== r && null !== r) {
                t.length && t.push("&");
                var a = encodeURIComponent("boolean" == typeof r ? (r ? "1" : "0") : r.toString());
                t.push(encodeURIComponent(n), "=", a);
              }
            }
            return t.join("");
          }
          return "string" == typeof e ? e : null;
        },
        s = function (e, t) {
          if (((t = o(t)), "string" == typeof t)) {
            var n = /\?/g.test(e);
            e += (n ? "&" : "?") + o(t);
          }
          return e;
        },
        u = function (e, t, n) {
          (e.setAttribute("type", "text/javascript"),
            n && e.setAttribute("charset", n),
            e.setAttribute("src", t),
            document.getElementsByTagName("head")[0].appendChild(e));
        },
        l = function (e) {
          if (e.clearAttributes) e.clearAttributes();
          else for (var t in e) e.hasOwnProperty(t) && delete e[t];
          (e && e.parentNode && e.parentNode.removeChild(e), (e = null));
        },
        p = function (e, t, n) {
          function r(e) {
            return function () {
              try {
                (e ? p.onfailure && p.onfailure() : (t.apply(window, arguments), clearTimeout(i)),
                  (window[a] = null),
                  delete window[a]);
              } catch (n) {
              } finally {
                l(o);
              }
            };
          }
          var a,
            i,
            o = document.createElement("SCRIPT"),
            s = "bd__cbs__",
            p = n || {},
            f = p.charset,
            d = p.queryField || "callback",
            g = p.timeOut || 0,
            h = new RegExp("(\\?|&)" + d + "=([^&]*)");
          ((a = c.getUniqueId(s)),
            (window[a] = r(0)),
            g && (i = setTimeout(r(1), g)),
            (e = e.replace(h, "$1" + d + "=" + a)),
            e.search(h) < 0 && (e += (e.indexOf("?") < 0 ? "?" : "&") + d + "=" + a),
            u(o, e, f));
        },
        f = function (e, t) {
          var n = [];
          (n.push("<form id='", a, "' target='", i, "' "), n.push("action='", c.encodeHTML(e), "' method='post'>"));
          var r = {};
          for (var o in t)
            if (t.hasOwnProperty(o)) {
              var s = t[o];
              if (void 0 !== s && null !== s) {
                var u = "boolean" == typeof s ? (s ? "1" : "0") : s;
                r[o] = u;
              }
            }
          try {
            var l = "OOOO00",
              p = "OOO00O",
              f = "OOO000",
              d = "OOO0OO",
              g = "O0OOO0",
              h = { OOOOO0: l, O00000: p, O0O00O: f, O000OO: d, O0O000: g },
              m = new Date().getTime() / 1e3,
              v = parseInt(m / 86400, 10) % 5,
              w = [];
            if (Object.keys) w = Object.keys(h);
            else {
              w = [];
              for (var O in h) w.push(O);
            }
            var y = h[w[v]] || "";
            window.moonshadV3 && window.moonshadV3[y] && r && (r = baidu.extend(r, window.moonshadV3[y](r)));
          } catch (I) {}
          for (var o in r)
            r.hasOwnProperty(o) &&
              n.push("<input type='hidden' name='", c.encodeHTML(o), "' value='", c.encodeHTML(r[o]), "' />");
          return (n.push("</form>"), n.join(""));
        },
        d = function (e, t, n, o) {
          function s(e) {
            return function () {
              try {
                (e ? o.onfailure && o.onfailure() : (n.apply(window, arguments), l && clearTimeout(l)),
                  (window[p] = null),
                  delete window[p]);
              } catch (t) {}
            };
          }
          o = o || {};
          var u = o.timeOut || 0,
            l = !1,
            p = c.getUniqueId("bd__pcbs__");
          t[o.queryField || "callback"] = "parent." + p;
          var d = f(e, t);
          if (c.g(a)) c.getParent(a).innerHTML = d;
          else {
            var g = [];
            (g.push("<div id='", r, "' style='display:none;'>"),
              g.push("<div>", d, "</div>"),
              g.push(
                "<iframe allow='local-network-access *' name='",
                i,
                "' src='" +
                  ("https:" ===
                  (window.location ? window.location.protocol.toLowerCase() : document.location.protocol.toLowerCase())
                    ? "https://passport.baidu.com/passApi/html/_blank.html"
                    : "about:blank") +
                  "' style='display:none;'></iframe>",
              ),
              g.push("</div>"),
              c.insertHTML(document.body, "beforeEnd", g.join("")));
          }
          ((window[p] = s()), u && (l = setTimeout(s(1), u)), c.g(a).submit());
        };
      ((n.jsonp = function (n, r, a) {
        a = a || {};
        ((r.lang = passLang), e && e.traceID && e.traceID.createTraceID && (r.traceid = e.traceID.createTraceID()));
        var i = {};
        for (var o in r) {
          var c = r[o];
          void 0 !== c && null !== c && (i[o] = c);
        }
        try {
          var u = "OOOO00",
            l = "OOO00O",
            f = "OOO000",
            d = "OOO0OO",
            g = "O0OOO0",
            h = { OOOOO0: u, O00000: l, O0O00O: f, O000OO: d, O0O000: g },
            m = new Date().getTime() / 1e3,
            v = parseInt(m / 86400, 10) % 5,
            w = [];
          if (Object && Object.keys) w = Object.keys(h);
          else {
            w = [];
            for (var O in h) w.push(O);
          }
          var y = h[w[v]] || "";
          window.moonshadV3 && window.moonshadV3[y] && r && (r = baidu.extend(r, window.moonshadV3[y](i, baidu)));
        } catch (I) {}
        return new t(function (e, t) {
          ((n = s(n, r)),
            p(
              n,
              function (t) {
                (a.processData && (t = a.processData(t)), e && e(t));
              },
              {
                charset: a.charset,
                queryField: a.queryField,
                timeOut: a.timeOut,
                onfailure: function () {
                  t && t();
                },
              },
            ));
        });
      }),
        (n.submit = function (n, r, a) {
          return (
            e && e.traceID && e.traceID.createTraceID && (r.traceid = e.traceID.createTraceID()),
            (r.lang = passLang),
            n && r
              ? new t(function (e) {
                  d(
                    n,
                    r,
                    function (t) {
                      (a.processData && (t = a.processData(t)), e && e(t));
                    },
                    a,
                  );
                })
              : void 0
          );
        }));
      var g = [];
      n.load = function (e) {
        return new t(function (t) {
          var n = g.push(new Image()) - 1,
            r = !1,
            a = setTimeout(function () {
              ((r = !0), t && t());
            }, 1e3);
          ((g[n].onload = function () {
            (clearTimeout(a), r || (t && t()), (r = !0), (g[n] = g[n].onload = null));
          }),
            (g[n].src = e));
        });
      };
    })(u);
    var l = passport.requestDomain || "https://passport.baidu.com",
      p = {
        getApiInfo: "/v2/api/?getapi",
        getLoginHistory: "/v2/api/?loginhistory",
        loginCheck: "/v2/api/?logincheck",
        getVerifyCodeStr: "/v2/?reggetcodestr",
        getRegSmsVerifyCodeStr: "/v2/?regsmscodestr",
        checkUserName: "/v2/?regnamesugg",
        checkPassword: "/v2/?regpwdcheck",
        checkMail: "/v2/?regmailcheck",
        isUserNoName: "/v2/api/?ucenteradduname",
        checkPhone: "/v2/?regphonecheck",
        getphonestatus: "/v2/?getphonestatus",
        sendPhoneCode: "/v2/?regphonesend",
        multiBind: "/v2/?multiaccountassociate",
        multiUnbind: "/v2/?multiaccountdisassociate",
        multiCheckUserName: "/v2/?multiaccountusername",
        multiGetaccounts: "/v2/?multiaccountget",
        multiSwitchuser: "/v2/?loginswitch",
        checkVerifycode: "/v2/?checkvcode",
        getRsaKey: "/v2/getpublickey",
        authwidGetverify: "/v2/sapi/authwidgetverify",
        checkIDcard: "/v3/finance/main/idnumcert",
        checkIDcardSecondStep: "/v3/finance/main/upcert",
        checkIDcardAllStep: "/v3/finance/main/idnumcert",
        checkIDcardState: "/v3/finance/main/checkupcert",
        syncIntBDUSS: "/v3/api/login/syncintbduss",
        syncPhone: "/v3/api/conf/getupsms",
        unbindingSms: "/v3/api/user/wlwforbidfree",
        certUpLevel: "/v3/api/fraud/liftban",
      },
      f = {
        login: "/v2/api/?login",
        reg: "/v2/api/?reg",
        fillUserName: "/v2/api/?ucenteradduname",
        regPhone: "/v2/api/?regphone",
        checkIDcard: "/v3/finance/main/idnumcert",
        checkIDcardSecondStep: "/v3/finance/main/upcert",
        checkIDcardAllStep: "/v3/finance/main/idnumcert",
        getAccessToken: "/v3/login/api/getoauthtoken",
        wxGetUserNameInit: "/v8/portal/api/username_sugg_init",
      },
      d = {
        getApiInfo: { apiType: "class" },
        login: {
          memberPass: "mem_pass",
          safeFlag: "safeflg",
          isPhone: "isPhone",
          timeSpan: "ppui_logintime",
          logLoginType: "logLoginType",
          isOver: "isOver",
        },
        fillUserName: {
          selectedSuggestName: "pass_fillinusername_suggestuserradio",
          timeSpan: "ppui_fillusernametime",
        },
        reg: {
          password: "loginpass",
          timeSpan: "ppui_regtime",
          suggestIndex: "suggestIndex",
          suggestType: "suggestType",
          selectedSuggestName: "pass_reg_suggestuserradio_0",
          logRegType: "logRegType",
        },
        regPhone: {
          password: "loginpass",
          timeSpan: "ppui_regtime",
          suggestIndex: "suggestIndex",
          suggestType: "suggestType",
          selectedSuggestName: "pass_reg_suggestuserradio_0",
          logRegType: "logRegType",
        },
      },
      g = {
        loginCheck: {
          isPhone: function (e) {
            return e ? "true" : "false";
          },
        },
        login: {
          memberPass: function (e) {
            return e ? "on" : "";
          },
        },
      },
      h = {
        checkPassword: { fromreg: 1 },
        reg: {
          registerType: 1,
          verifypass: function (e) {
            return e.password;
          },
        },
      },
      m = { password: !0 },
      v = { login: function () {} },
      w = {
        checkUserName: "reg",
        checkMail: "reg",
        checkPhone: "regPhone",
        sendPhoneCode: "regPhone",
        multiCheckUserName: "multiBind",
        multiSwitchuser: "changeUser",
        checkVerifycode: "checkVerifycode",
      },
      O = passport.err.getCurrent().errMsg || passport.err.getCurrent(),
      y = {};
    ((e.setContext = function (e) {
      ((y.product = e.product || y.product),
        (y.charset = e.charset || y.charset),
        (y.staticPage = e.staticPage || y.staticPage),
        (y.token = e.token || y.token),
        (y.subpro = e.subpro || y.subpro));
    }),
      (e.traceID = {
        headID: (e.traceID && e.traceID.headID) || "",
        flowID: (e.traceID && e.traceID.flowID) || "",
        cases: (e.traceID && e.traceID.cases) || "",
        initTraceID: function (e) {
          var t = this;
          e && e.length > 0 ? ((t.headID = e.slice(0, 6)), (t.flowID = e.slice(6, 8))) : t.destory();
        },
        createTraceID: function () {
          var e = this;
          return e.headID + e.flowID + e.cases;
        },
        startFlow: function (e) {
          var t = this,
            n = t.getFlowID(e);
          0 === t.flowID.length || t.flowID === n ? (t.createHeadID(), (t.flowID = n)) : t.finishFlow(n);
        },
        finishFlow: function () {
          var e = this;
          e.destory();
        },
        getRandom: function () {
          return parseInt(90 * Math.random() + 10, 10);
        },
        createHeadID: function () {
          var e = this,
            t = new Date().getTime() + e.getRandom().toString(),
            n = Number(t).toString(16),
            r = n.length,
            a = n.slice(r - 6, r).toUpperCase();
          e.headID = a;
        },
        getTraceID: function (e) {
          var t = this,
            n = (e && e.traceid) || "";
          t.initTraceID(n);
        },
        getFlowID: function (e) {
          var t = { login: "01", reg: "02" };
          return t[e];
        },
        setData: function (e) {
          var t = this;
          return (
            e.data
              ? (e.data.traceid = t.createTraceID())
              : (e.url = e.url + (e.url.indexOf("?") > -1 ? "&" : "?") + "traceid=" + t.createTraceID()),
            e
          );
        },
        destory: function () {
          var e = this;
          ((e.headID = ""), (e.flowID = ""));
        },
      }));
    for (var I in p) p.hasOwnProperty(I) && (e[I] = n(I, p[I]));
    for (var I in f) f.hasOwnProperty(I) && (e[I] = n(I, f[I], !0));
    ((e.jsonp = function (e, t) {
      return (
        0 != e.indexOf("http") && (e = l + e),
        (t = t || {}),
        (t.flag_code && 1 == t.flag_code) || (t.apiver = "v3"),
        (t.tt = new Date().getTime()),
        u.jsonp(e, t, {
          charset: "utf-8",
          processData: function (e) {
            return o(e);
          },
        })
      );
    }),
      (e.post = function (e, t) {
        return (
          (t = t || {}),
          "wap" !== t.apitype && (e = l + e),
          (t.staticpage = t.staticpage || y.staticPage),
          (t.charset = t.charset || y.charset || document.characterSet || document.charset || ""),
          (t.token = t.token || y.token),
          (t.tpl = t.tpl || y.product),
          u.submit(e, t, {
            charset: "utf-8",
            processData: function (e) {
              return o(e);
            },
          })
        );
      }),
      (e.request = u));
  })(passport.data);
  var passport = passport || window.passport || {};
  ((passport.analysis = passport.analysis || {}),
    (function (e) {
      var i = function (e, i) {
        var o = e.config.diaPassLogin ? "dialogLogin" : "basicLogin",
          n = e.config.loginMerge ? 1 : 0,
          t = e.config.product || "isnull",
          a = window.location ? window.location.protocol.toLowerCase() : document.location.protocol.toLowerCase(),
          r = "",
          s = "&tt=" + new Date().getTime(),
          l = e.guideRandom ? e.guideRandom : "";
        for (var g in i) r = r + "&" + g + "=" + i[g];
        if ("http:" == a)
          var c =
            "http://nsclick.baidu.com/v.gif?pid=111&url=&logintype=" +
            o +
            "&gid=" +
            l +
            "&merge=" +
            n +
            "&tpl=" +
            t +
            r +
            s;
        else if ("https:" == a)
          var c =
            "https://passport.baidu.com/img/v.gif?logintype=" + o + "&gid=" + l + "&merge=" + n + "&tpl=" + t + r + s;
        if (c) {
          var p = new Image();
          ((p.onload = p.onerror =
            function () {
              ((p.onload = p.onerror = null), (p = null));
            }),
            (p.src = c));
        }
      };
      e.login = {
        render: function (e) {
          (i(e, { type: "firstrender", loginurl: encodeURIComponent(document.location.href) }),
            baidu(e.getPhoenixElement("pass_phoenix_list_login")).on("click", function (o) {
              var n,
                t = baidu(o.target);
              if (t && t.attr("title")) {
                switch (t.attr("title")) {
                  case "普通登录":
                    n = "normal";
                    break;
                  case "二维码登录":
                    n = "qrcode";
                    break;
                  case "短信登录":
                    n = "sms";
                    break;
                  case "QQ帐号":
                    n = "qq";
                    break;
                  case "新浪微博":
                    n = "weibo";
                    break;
                  case "人人网":
                    n = "renren";
                    break;
                  case "腾讯微博":
                    n = "tqq";
                    break;
                  case "飞信":
                    n = "fetion";
                    break;
                  case "微信":
                    n = "weixin";
                    break;
                  case "钉钉":
                    n = "dingtalk";
                    break;
                  case "苹果":
                    n = "apple";
                    break;
                  case "天翼":
                    n = "tianyi";
                }
                i(e, { phoenix: n });
              }
            }));
          var o = (e.getElement(), e.getElement("form"));
          baidu(o).on("submit", function () {
            e.loginfirstsubmit || ((e.loginfirstsubmit = !0), i(e, { type: "loginfirstsubmit" }));
          });
        },
        changeLoginType: function (e, o) {
          i(e, { type: "changelogintype", logintype: (o && o.loginType) || "" });
        },
        fieldFocus: function (e, o) {
          (o.ele.get(0).id != e.$getId("smsPhone") && o.ele.get(0).id != e.$getId("smsVerifyCode")) ||
          e.smsloginfirstlog
            ? e.loginfirstlog || ((e.loginfirstlog = !0), i(e, { type: "loginfirst" }))
            : ((e.smsloginfirstlog = !0), i(e, { type: "smsloginfirst" }));
        },
        loginSuccess: function (e) {
          i(e, { type: "loginsuccess" });
        },
        loginError: function () {},
        validateError: function (e, o) {
          return (
            o.validate && i(e, { errno: encodeURIComponent(o.validate.msg), type: "loginerrno" }),
            { preventEvent: !1, preventDefault: !1 }
          );
        },
        fieldKeyup: function (e) {
          e.KEYUPFLAG || (i(e, { type: "typein" }), (e.KEYUPFLAG = !0));
        },
      };
    })(passport.analysis));
  window.passport.sendAjax = function (obj) {
    try {
      var url = obj.url;
      var method = obj.method || "get";
      var async = obj.async === undefined ? true : obj.async;
      var data = obj.data || {};
      function splitStr(data) {
        let str = "";
        for (var key in data) {
          let s = key + "=" + data[key] + "&";
          str += s;
        }
        return str.substring(0, str.length - 1);
      }
      // initial
      var xhttp;
      if (window.XMLHttpRequest) {
        xhttp = new XMLHttpRequest();
      } else {
        xhttp = new ActiveXObject();
      }
      xhttp.timeout = obj.timeout || 100000;
      // send
      if (method.toLowerCase() === "get") {
        if (data === undefined || JSON.stringify(data) === "{}") {
          xhttp.open("GET", url, async);
          xhttp.send();
        } else {
          xhttp.open("GET", url + "?" + splitStr(data), async);
          xhttp.send();
        }
      } else if (method.toLowerCase() === "post") {
        if (data === undefined) {
          data = {};
        }
        xhttp.open("POST", url, async);
        xhttp.withCredentials = true; // 支持跨域发送cookies
        xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        xhttp.send(splitStr(data));
      }
      // response
      xhttp.onreadystatechange = function () {
        if (xhttp.readyState === 4) {
          if (xhttp.status === 200 || xhttp.status === 304) {
            try {
              obj.success(JSON.parse(xhttp.responseText));
            } catch (error) {
              obj.success(xhttp.responseText);
            }
          } else {
            obj.error();
          }
        }
      };
    } catch (error) {
      obj.error();
    }
  };

  /**
   * 批量请求客户端互通接口
   * @param {*} clientShareOpt，客户端互通接口的参数
   * @param {*} successCallback 请求成功后的回调
   * @param {*} errorCallback 所有请求失败回调
   */
  window.passport.tryAllClientSharePorts = function (clientShareOpt, successCallback, errorCallback) {
    try {
      var shareOrigin = clientShareOpt.shareOrigin || "";
      var sharePortArr = clientShareOpt.sharePortArr || [];
      var getClientTokenApi = clientShareOpt.getClientTokenApi || "";
      var customParmasObj = clientShareOpt.customParmasObj || {};
      var currentIndex = 0;
      function fetchToken() {
        if (currentIndex >= sharePortArr.length) {
          // 所有端口都尝试过了
          errorCallback(new Error("All ports failed to fetch token."));
          return;
        }
        var port = sharePortArr[currentIndex];
        var url = shareOrigin + ":" + port + getClientTokenApi;

        // 发送请求
        window.passport.sendAjax({
          url: url,
          method: "get",
          timeout: 1000,
          data: customParmasObj,
          success: function (response) {
            // 请求成功，调用 successCallback
            successCallback(response);
          },
          error: function () {
            // 端口失败，尝试下一个端口
            currentIndex++;
            fetchToken();
          },
        });
      }
      // 支持Promise情况，走并发，不支持走串行
      if (window.Promise && window.Promise.any) {
        function newSendAjaxPromise(url) {
          return new Promise(function (resolve, reject) {
            window.passport.sendAjax({
              url: url,
              method: "get",
              timeout: 1000,
              data: customParmasObj,
              success: function (response) {
                resolve(response);
              },
              error: function () {
                reject();
              },
            });
          });
        }
        var promiseArr = [];
        for (var i = 0; i < sharePortArr.length; i++) {
          var port = sharePortArr[i];
          var url = shareOrigin + ":" + port + getClientTokenApi;
          promiseArr.push(newSendAjaxPromise(url));
        }
        Promise.any(promiseArr)
          .then(function (response) {
            successCallback(response);
          })
          .catch(function (error) {
            errorCallback(new Error("All ports failed to fetch token."));
          });
      } else {
        // 开始尝试第一个端口
        fetchToken();
      }
    } catch (error) {
      errorCallback(new Error("Failed to initialize the request"));
    }
  };
  var passport = passport || window.passport || {};
  ((passport.hook = passport.hook || {}),
    (function (e) {
      function n(e) {
        var n,
          o,
          a = this,
          t = {
            120016: { isLogin: !1, msg: A.bindAction16 },
            400032: { isLogin: !0, msg: A.bindAction32 },
            400034: { isLogin: !1, msg: { phone: A.bindAction34, email: A.bindAction } },
          }[e.errno],
          r = e.args,
          i = e.title,
          s = t.msg,
          c = (e.auth_title, e.auth_msg, t.isLogin),
          l = e.cfg,
          p = function (e, n, o) {
            var a = n.args,
              t = {
                action: n.type || "init",
                u: e.config.u,
                tpl: e.config.product,
                ltoken: a.rsp.data.ltoken,
                lstr: a.rsp.data.lstr,
              };
            ((e.REQUESTBINDTOKENURL = "/v2/?loginspmbindsecureinfo"),
              passport.data.jsonp((passport.requestDomain || M) + e.REQUESTBINDTOKENURL, t).success(function (n) {
                0 == n.errInfo.no
                  ? o &&
                    o({
                      bindEmailToken: n.data.bindEmailToken,
                      bindMobileToken: n.data.bindMobileToken,
                      authsid: n.data.authsid,
                      loginproxy: n.data.loginproxy,
                      otherValue: t,
                    })
                  : (e._ownerDialog && e._ownerDialog.show(), (e.getElement("error").innerHTML = e.lang.sysError));
              }));
          },
          d = function (e) {
            var n = "string" == (typeof s).toLowerCase() ? s : s.email;
            return (
              (n += c
                ? W.youCan +
                  '<a class="bindLink bindJumpEmail">' +
                  W.jump +
                  "</a>" +
                  W.or +
                  '<a class="bindLink bindPhoneBtn">' +
                  W.bindPhone +
                  "</a>。"
                : W.youAlsoCan + '<a class="bindLink bindPhoneBtn">' + W.bindPhone + "</a>。"),
              (e.otherValue = e.otherValue || {}),
              passport.pop.ArmorWidget("bindemail", {
                token: e.bindEmailToken,
                authsid: e.authsid,
                title: i || W.bindPassMail,
                otherValue: e.otherValue,
                msg: n,
                lstr: e.otherValue.lstr || "",
                ltoken: e.otherValue.ltoken || "",
                u: e.otherValue.u || "",
                tpl: e.otherValue.tpl || "",
                subpro: a.config.subpro,
                traceid: e.traceid,
                onSubmitSuccess: function (e) {
                  var e = e;
                  p(a, { args: r, type: "check" }, function (n) {
                    (e && e.hide && e.hide(),
                      (r.isCompleted = !0),
                      n.loginproxy
                        ? passport.data.jsonp(n.loginproxy).success(function (e) {
                            l.onCompleted &&
                              l.onCompleted(e, function () {
                                l.onCancel && l.onCancel(r);
                              });
                          })
                        : l.onCancel && l.onCancel(r));
                  });
                },
                onRender: function () {
                  var e = this;
                  (baidu(".bindPhoneBtn").on("click", function () {
                    (e.close(), o.show());
                  }),
                    baidu(".bindJumpEmail").on("click", function (n) {
                      (n.preventDefault(), e.close(), (r.isCompleted = !0), l.onCancel && l.onCancel(r));
                    }),
                    baidu("#" + e.getId("header_a")).on("click", function () {
                      c && ((r.isCompleted = !0), l.onCancel && l.onCancel(r));
                    }));
                },
              })
            );
          },
          u = function (e) {
            var o = "string" == (typeof s).toLowerCase() ? s : s.phone;
            return (
              c && e.bindEmailToken
                ? (o +=
                    W.youCan +
                    '<a class="bindLink bindJumpPhone">' +
                    W.jump +
                    "</a>" +
                    W.or +
                    '<a class="bindLink bindEmailBtn">' +
                    W.bindPassMail +
                    "</a>。")
                : e.bindEmailToken &&
                  (o += W.youAlsoCan + '<a class="bindLink bindEmailBtn">' + W.bindPassMail + "</a>。"),
              (e.otherValue = e.otherValue || {}),
              passport.pop.ArmorWidget("bindmobile", {
                token: e.bindMobileToken,
                authsid: e.authsid,
                title: i || W.bindPhone,
                otherValue: e.otherValue,
                msg: o,
                bindToLogin: 1,
                entryScene: "login",
                lstr: e.otherValue.lstr || "",
                ltoken: e.otherValue.ltoken || "",
                u: e.otherValue.u || "",
                tpl: e.otherValue.tpl || "",
                apiMargicInstance: a,
                subpro: a.config.subpro,
                traceid: e.traceid,
                staticPage: a.config.staticPage || "",
                onSubmitSuccess: function (e) {
                  var e = e;
                  p(a, { args: r, type: "check" }, function (n) {
                    (e && e.hide && e.hide(),
                      (r.isCompleted = !0),
                      n.loginproxy
                        ? passport.data.jsonp(n.loginproxy).success(function (e) {
                            l.onCompleted &&
                              l.onCompleted(e, function () {
                                l.onCancel && l.onCancel(r);
                              });
                          })
                        : l.onCancel && l.onCancel(r));
                  });
                },
                onRender: function (o) {
                  var o = this;
                  (baidu(".bindEmailBtn").on("click", function (a) {
                    (a.preventDefault(), o.close(), (n = n || d(e)), n.show());
                  }),
                    baidu(".bindJumpPhone").on("click", function (e) {
                      (e.preventDefault(), o.close(), (r.isCompleted = !0), l.onCancel && l.onCancel(r));
                    }),
                    baidu("#" + o.getId("header_a")).on("click", function () {
                      c && ((r.isCompleted = !0), l.onCancel && l.onCancel(r));
                    }));
                },
                onBindToLoginFn: function (e, n) {
                  n &&
                    n.mobile &&
                    (a.config.sms
                      ? (a.getElement("smsPhone_placeholder") && a.$hide("smsPhone_placeholder"),
                        a.getElement("smsPhone") && (a.getElement("smsPhone").value = n.mobile),
                        a.getElement("smsVerifyCode") &&
                          ((a.getElement("smsVerifyCode").value = ""), a.getElement("smsVerifyCode").focus()))
                      : (a.getElement("userName_placeholder") && a.$hide("userName_placeholder"),
                        a.getElement("userName") && (a.getElement("userName").value = n.mobile),
                        a.getElement("password") &&
                          ((a.getElement("password").value = ""), a.getElement("password").focus())));
                },
              })
            );
          };
        p(a, { args: r, type: "init" }, function (e) {
          passport._use(I, L[I], function () {
            e.bindMobileToken
              ? ((o = u(e)), o.show())
              : e.bindEmailToken
                ? ((n = d(e)), n.show())
                : (a._ownerDialog && a._ownerDialog.show(), (a.getElement("error").innerHTML = a.lang.sysError));
          });
        });
      }
      function o(e) {
        var n,
          o,
          a = this,
          t = e.rspData,
          r = e.cfg,
          i = e.args,
          s = function (e) {
            return (
              (e = e || A.getWarnHtml),
              '<div class="passport-forceverify-risk"><p class="passport-forceverify-risk-msg">' +
                e +
                '</p><div  class="passport-forceverify-risk-con clearfix"><a class="passport-forceverify-risk-next" id="passport_forceverify_risk_next" href="###">下次提醒</a><a class="passport-forceverify-risk-btn" href="http://passport.baidu.com/v2/account/password" target="_blank" >立即修改</a></div></div>'
            );
          };
        if (t && t.secstate)
          switch (t.secstate) {
            case "PA001":
              o = W.warnMsg1;
              break;
            case "PA002":
              o = W.warnMsg2;
              break;
            case "PA003":
              o = W.warnMsg3;
              break;
            case "risk":
              ((n = s()), (o = W.warnMsg4));
              break;
            case "cheat":
              o = W.warnMsg5;
              break;
            case "PC001":
              o = W.warnMsg6;
              break;
            case "PX008":
              o = W.warnMsg7;
              break;
            default:
              o = W.warnMsg8;
          }
        var c = {
          400031: { title: W.loginProtect, msg: W.loginTip1 },
          5: {
            title: W.loginSafeV,
            msg: W.warnMsg2,
            onSuccess: function (e, n) {
              var o = t.gotourl + "&authsid=" + n.authsid;
              passport.data.jsonp(o).success(function (n) {
                (a._ownerDialog && a._ownerDialog.show(), e.hide());
                var o = (n.errInfo && n.errInfo.no) || (n.data && n.data.errno);
                a.getElement("error").innerHTML =
                  0 !== +o || O
                    ? a.lang.sysError
                    : W.reLogin +
                      "，" +
                      W.or +
                      '<a href="https://passport.baidu.com/?getpass_index&tpl=' +
                      B +
                      "&lang=" +
                      R +
                      '" target="_blank">' +
                      W.findPassword +
                      "</a>";
              });
            },
            onGetapiError: function () {
              "sms" === (a.config && a.config.currentTab)
                ? (a.getElement("smsError").innerHTML = W.loginTip2)
                : (a.getElement("error").innerHTML = W.loginTip2);
            },
          },
          400023: {
            title: W.loginSafeV,
            msg: W.warnMsg2,
            onSuccess: function (e, n) {
              var o = (passport.requestDomain || M) + "/v3/login/main/qrbdusslogin?tt=" + new Date().getTime(),
                s = {
                  authsid: n.authsid,
                  bduss: t.bdusssign,
                  u: encodeURIComponent(t.u),
                  loginVersion: "v4",
                  tpl: a.config.product,
                };
              passport.data.jsonp(o, s).success(function (n) {
                (e.hide(),
                  0 == n.errInfo.no || 0 == n.data.errno
                    ? ((i.isCompleted = !0),
                      r.onCompleted &&
                        r.onCompleted(n, function () {
                          r.onCancel && r.onCancel(i);
                        }))
                    : (a.getElement("error").innerHTML = a.lang.sysError));
              });
            },
            onGetapiError: function () {
              a.getElement("error").innerHTML = W.loginTip2;
            },
          },
          120019: {
            title: W.loginUnfreeze,
            msg: W.loginTip3,
            onSuccess: function (e, n) {
              var o = t.gotourl + "&authsid=" + n.authsid;
              passport.data.jsonp(o).success(function (n) {
                (a._ownerDialog && a._ownerDialog.show(), e.hide());
                var o = (n.errInfo && n.errInfo.no) || (n.data && n.data.errno);
                a.getElement("error").innerHTML =
                  0 !== +o || O
                    ? a.lang.sysError
                    : W.reLogin +
                      "，" +
                      W.or +
                      '<a href="https://passport.baidu.com/?getpass_index&tpl=' +
                      B +
                      "&lang=" +
                      R +
                      '" target="_blank">' +
                      W.findPassword +
                      "</a>";
              });
            },
            onGetapiError: function () {
              a.getElement("error").innerHTML = O
                ? W.loginLimit
                : W.loginLimit +
                  "，" +
                  W.youCan +
                  '<a href="https://passport.baidu.com/?getpass_index&tpl=' +
                  B +
                  "&lang=" +
                  R +
                  '" target="_blank">' +
                  W.findPassword +
                  "</a>" +
                  W.or +
                  W.thressThour;
            },
          },
          120021: {
            title: W.safety,
            msg: o,
            defaultHTML: n,
            onSuccess: function (e) {
              return (
                passport.data.jsonp(t.loginproxy).success(function (o) {
                  (e.show(),
                    0 == o.errInfo.no
                      ? ((i.isCompleted = !0),
                        i.rsp && i.rsp.data && (i.rsp.data.hao123Param = o.data.hao123Param),
                        o.data && o.data.sync_id && _(o.data.sync_id, a),
                        n
                          ? ((e.getElement("article").innerHTML = n),
                            baidu(e.getElement("header_a")).on("click", function () {
                              (e.hide(),
                                r.onCompleted &&
                                  r.onCompleted(o, function () {
                                    r.onCancel && r.onCancel(i);
                                  }));
                            }),
                            baidu(document.getElementById("passport_forceverify_risk_next")).on("click", function () {
                              (e.hide(),
                                r.onCompleted &&
                                  r.onCompleted(o, function () {
                                    r.onCancel && r.onCancel(i);
                                  }));
                            }))
                          : (e.hide(),
                            r.onCompleted &&
                              r.onCompleted(o, function () {
                                r.onCancel && r.onCancel(i);
                              })))
                      : e.getElement("forceverify_error").html(a.lang.sysError));
                }),
                !1
              );
            },
            onRender: function (e) {
              document.getElementById("passport_forceverify_risk_appeal") &&
                (document.getElementById("passport_forceverify_risk_appeal").href = e.url_forgot);
            },
          },
          riskCheat: {
            token: "risk",
            title: W.safety,
            msg: o,
            defaultHTML: n,
            onRender: function (e) {
              (baidu(document.getElementById("passport_forceverify_risk_next")).on("click", function () {
                (e.hide(), r.onCancel && r.onCancel());
              }),
                document.getElementById("passport_forceverify_risk_appeal") &&
                  (document.getElementById("passport_forceverify_risk_appeal").href = e.url_forgot));
            },
          },
        }[e.errno];
        passport._use(H, L[H], function () {
          forceverifyLoginverify = passport.pop.Forceverify(
            {
              token: t.authtoken,
              title: c.title,
              msg: c.msg,
              subpro: a.config.subpro,
              u: t.u || "",
              lstr: t.lstr || "",
              ltoken: t.ltoken || "",
              tpl: t.tpl || "",
              traceid: t.traceid,
              lang: a.config.lang,
              onRender: function (e) {
                c.onRender && c.onRender(e);
              },
              onSubmitSuccess: function (e, n) {
                try {
                  "localStorage" in window &&
                    null !== window.localStorage &&
                    localStorage.setItem("passLoginType", "normal");
                } catch (o) {}
                if ("1" === t.switchtobind && "1" === t.realnameverifyemail)
                  return ((t.bindwidgetauthsid = n.authsid), void h.apply(a, [{ args: i, rspData: t, cfg: r }]));
                if ("1" === t.realnameverifyemail)
                  return ((t.realnameauthsid = n.authsid), void g.apply(a, [{ args: i, rspData: t, cfg: r }]));
                if (c.onSuccess) return void c.onSuccess(e, n);
                var s = t.loginproxy;
                passport.data.jsonp(s).success(function (n) {
                  if (0 == n.errInfo.no || 0 == n.data.errno) {
                    if (
                      (e.hide(),
                      (i.isCompleted = !0),
                      i.rsp && i.rsp.data && ((i.rsp.data.hao123Param = n.data.hao123Param), b(i, n)),
                      n.data && n.data.sync_id && _(n.data.sync_id, a),
                      n.data && n.data.secret_id)
                    ) {
                      var o = i.rsp.data.u.indexOf("?") > 0 ? "&" : "?";
                      ((i.rsp.data.u += o + "secret_id=" + n.data.secret_id),
                        (i.rsp.data.secret_id = n.data.secret_id || ""));
                    }
                    r.onCompleted &&
                      r.onCompleted(n, function () {
                        r.onCancel && r.onCancel(i);
                      });
                  } else
                    (e.hide(),
                      a._ownerDialog && a._ownerDialog.show(),
                      (a.getElement("error").innerHTML = a.lang.sysError));
                });
              },
              onGetapiError: function (e) {
                return (
                  a._ownerDialog && a._ownerDialog.show(),
                  c.onGetapiError ? void c.onGetapiError(e) : void (a.getElement("error").innerHTML = a.lang.sysError)
                );
              },
              onHide: function () {
                (r.onCancel && r.onCancel(),
                  a.config.loginMerge && a.getElement("isPhone") && (a.getElement("isPhone").value = ""));
              },
            },
            !0,
          );
        });
      }
      function a(e) {
        if (e) {
          var n = e.getElement("submit");
          n && ((n.disabled = !1), (n.value = (e.lang && e.lang.login) || ""), n.style && (n.style.color = "#fff"));
        }
      }
      function t(e) {
        function n() {
          var e = document.getElementById("managePcAuthwidget") || "";
          e && document.body.removeChild(e);
        }
        var o = this,
          t = e.rspData,
          r = e.cfg,
          i = e.args,
          s = "login_force_verify";
        400031 === +e.errno
          ? (s = "login_protect_verify")
          : 25 === +e.errno
            ? (s = "realname_cert")
            : 400425 === +e.errno && (s = "login_force_un_second_card_verify");
        var c = document.getElementById("managePcAuthwidget") || "";
        if (!c) {
          var l = document.createElement("div");
          ((l.id = "managePcAuthwidget"), (l.className = "pass-auth-widget-wrapper"), document.body.appendChild(l));
        }
        var p = !1,
          d = setTimeout(function () {
            p || ((p = !0), a(o));
          }, 6e3);
        passport._load(
          location.protocol +
            "//wappass.baidu.com/static/apimanage/api/authwidget.js?cdnversion=" +
            new Date().getTime(),
          !0,
          function () {
            var e;
            if (((p = !0), a(o), d && clearTimeout(d), window.PassApi && window.PassApi.authwidget)) {
              var c = window.PassApi.authwidget;
              e = new c({
                id: "managePcAuthwidget",
                mode: "dialog",
                authid: t.authid || "",
                pageType: "authType",
                scene: s,
                clientfrom: "pc",
                tpl: t.tpl || (passport && passport.tpl),
                staticPage: o.config.staticPage || "",
                lang: o.config.lang || "",
                verifySuccess: function (e, a) {
                  try {
                    "localStorage" in window &&
                      null !== window.localStorage &&
                      localStorage.setItem("passLoginType", "normal");
                  } catch (s) {}
                  if ("1" === t.switchtobind && "1" === t.realnameverifyemail)
                    return (n(), (t.bindwidgetauthsid = a), void h.apply(o, [{ args: i, rspData: t, cfg: r }]));
                  if ("1" === t.realnameswitch)
                    return (n(), (t.realnameauthsid = a), void g.apply(o, [{ args: i, rspData: t, cfg: r }]));
                  var c = t.loginproxy;
                  passport.data.jsonp(c).success(function (e) {
                    (n(),
                      0 == e.errInfo.no || 0 == e.data.errno
                        ? ((i.isCompleted = !0),
                          i.rsp && i.rsp.data && ((i.rsp.data.hao123Param = e.data.hao123Param), b(i, e)),
                          e.data && e.data.sync_id && _(e.data.sync_id, o),
                          r.onCompleted &&
                            r.onCompleted(e, function () {
                              r.onCancel && r.onCancel(i);
                            }))
                        : (o._ownerDialog && o._ownerDialog.show(),
                          (o.getElement("error").innerHTML = o.lang.sysError)));
                  });
                },
                verifyFail: function () {},
                closeAuthwidget: function () {
                  (o._ownerDialog && o._ownerDialog.show(),
                    n(),
                    r.onCancel && r.onCancel(),
                    o.config.loginMerge && o.getElement("isPhone") && (o.getElement("isPhone").value = ""));
                },
              });
            }
          },
        );
      }
      function r(e) {
        for (var n = this, o = e.rspData, a = [], t = o.accounts.split(";"), r = 0; r < t.length; r++) {
          var i = t[r].split(",");
          a.push({ username: i[0], portrait: i[1] });
        }
        passport._load(M + L[S], !0, function () {
          var e = passport.pop.init({
            type: "loginMultichoice",
            tangram: !0,
            apiOpt: { phone: o.userName, userList: a },
            onChoicedUser: function (o, a) {
              (e.hide(),
                n._ownerDialog && n._ownerDialog.show(),
                (n.getElement("userName").value = a.username),
                (n.getElement("isPhone").value = "false"),
                n.config.loginMerge && n.getElement("loginMerge") && (n.getElement("loginMerge").value = ""),
                "sms" == n.currentLoginType
                  ? ((n.getElement("smsHiddenFields_switchuname").value = a.username), n._submitSmsForm())
                  : (n.config.loginMerge || (n._collectData(), n.switchTo("normal"), n._restoreData("phone")),
                    n.submit()));
            },
            onHide: function () {
              ((n.getElement("userName").value = o.userName), (n.getElement("isPhone").value = ""));
            },
          });
          e.show();
        });
      }
      function i(e) {
        var n = this,
          o = n.$getId("pass_b2c_swf"),
          a = e.args.rsp.data,
          t = q(),
          r = 0;
        a && a.bckv && (r = parseInt(a.bckv, 10) > 0);
        var i = null;
        if (
          ((passport.CONSTANT = passport.CONSTANT || {}),
          (passport.CONSTANT.b2c_getlogin = function () {}),
          (passport.CONSTANT.b2c_setlogin = function () {
            var e = { kv: a.bckv, sync: a.bcsync, checksum: a.bcchecksum, time: a.bctime };
            try {
              if (i) {
                var n = i.get_movie(o, "b2c_setlogin").b2c_setlogin(e);
                try {
                  var t = document.createElement("script");
                  ((t.type = "text/javascript"),
                    (t.src = [
                      M,
                      "/v2/b2c-stable?",
                      "from=setlogin.done",
                      "&checksum=",
                      e.checksum,
                      "&time=",
                      e.time,
                      "&status=",
                      encodeURIComponent(n),
                    ].join("")),
                    document.getElementsByTagName("head")[0].appendChild(t));
                } catch (r) {}
              }
            } catch (r) {
              try {
                var t = document.createElement("script");
                ((t.type = "text/javascript"),
                  (t.src = [M, "/v2/b2c-stable?", "from=setlogin.fail", "&msg=", encodeURI(r.message)].join("")),
                  document.getElementsByTagName("head")[0].appendChild(t));
              } catch (r) {}
            }
          }),
          n.getElement("pass_b2c") && t.isexists && r)
        ) {
          var s = function (e) {
            var n = function (e) {
                e = e || {};
                var n,
                  o,
                  a,
                  t,
                  r = {},
                  i = function (e) {
                    return e
                      .replace(/&/g, "&amp;")
                      .replace(/</g, "&lt;")
                      .replace(/>/g, "&gt;")
                      .replace(/"/g, "&quot;")
                      .replace(/'/g, "&#39;");
                  };
                for (o in e) r[o] = e[o];
                e = r;
                var s = e.vars,
                  c = ["classid", "codebase", "id", "width", "height", "align"];
                if (
                  ((e.align = e.align || "middle"),
                  (e.classid = "clsid:d27cdb6e-ae6d-11cf-96b8-444553540000"),
                  (e.codebase =
                    "https://fpdownload.macromedia.com/pub/shockwave/cabs/flash/swflash.cab#version=6,0,0,0"),
                  (e.movie = e.url || ""),
                  delete e.vars,
                  delete e.url,
                  "string" == typeof s)
                )
                  e.flashvars = s;
                else {
                  var l = [];
                  for (o in s) ((t = s[o]), l.push(o + "=" + encodeURIComponent(t)));
                  e.flashvars = l.join("&");
                }
                var p = ["<object "];
                for (n = 0, a = c.length; a > n; n++) ((t = c[n]), p.push(" ", t, '="', i(e[t]), '"'));
                p.push(">");
                var d = {
                  wmode: 1,
                  scale: 1,
                  quality: 1,
                  play: 1,
                  loop: 1,
                  menu: 1,
                  salign: 1,
                  bgcolor: 1,
                  base: 1,
                  allowscriptaccess: 1,
                  allownetworking: 1,
                  allowfullscreen: 1,
                  seamlesstabbing: 1,
                  devicefont: 1,
                  swliveconnect: 1,
                  flashvars: 1,
                  movie: 1,
                };
                for (o in e)
                  ((t = e[o]),
                    (o = o.toLowerCase()),
                    d[o] && (t || t === !1 || 0 === t) && p.push('<param name="' + o + '" value="' + i(t) + '" />'));
                ((e.src = e.movie),
                  (e.name = e.id),
                  delete e.id,
                  delete e.movie,
                  delete e.classid,
                  delete e.codebase,
                  (e.type = "application/x-shockwave-flash"),
                  (e.pluginspage = "https://www.macromedia.com/go/getflashplayer"),
                  p.push("<embed"));
                var u;
                for (o in e)
                  if (((t = e[o]), t || t === !1 || 0 === t)) {
                    if (new RegExp("^salign$", "i").test(o)) {
                      u = t;
                      continue;
                    }
                    p.push(" ", o, '="', i(t), '"');
                  }
                return (u && p.push(' salign="', i(u), '"'), p.push("></embed></object>"), p.join(""));
              },
              o = function (e) {
                e = e || {};
                var o = document.createElement("div");
                ((o.innerHTML = n(e)),
                  (o.style.display = "none"),
                  document.getElementsByTagName("body")[0].appendChild(o));
              },
              a = /msie (\d+\.\d+)/i.test(navigator.userAgent) ? document.documentMode || +RegExp.$1 : void 0;
            return (
              (a > 7 || !a) && o(e),
              {
                get_movie: function (e, n) {
                  var o = document[e];
                  if (9 === document.documentMode && o && o.length)
                    for (var a = o.length, t = 0; a > t; t++) {
                      var r = o[t];
                      if ("embed" === r.tagName.toLowerCase()) {
                        o = r;
                        break;
                      }
                    }
                  return (n && "function" != typeof o[n] && (o = document.getElementById(e)), o);
                },
              }
            );
          };
          try {
            var c = window.location.protocol.toLowerCase();
            i = s({
              id: o,
              width: "1",
              height: "1",
              url: c + "//passport.baidu.com/passApi/swf/b2c.swf?_t=" + Math.random(),
              allowscriptaccess: "always",
            });
          } catch (l) {}
        }
        passport.data.request.load(
          M +
            "/v2/b2c-flash?isexists=" +
            encodeURIComponent(t.isexists) +
            "&ver=" +
            encodeURIComponent(t.ver) +
            "&filename=" +
            encodeURIComponent(t.filename),
        );
      }
      function s(e) {
        var n = this,
          o = e.rspData,
          a = e.args.rsp.data;
        return n.rebindGuideWidget
          ? void n.rebindGuideWidget.show()
          : void passport.use("uni_rebindGuide", { tangram: !0 }, function () {
              ((n.rebindGuideWidget = new passport.pop.rebindGuideWidget({
                color: n.config.color || "blue",
                apiOpt: a,
                rebindType: "rebindPhone",
                onrebindGuideCompleted: function () {
                  V(n, o);
                },
              })),
                n.rebindGuideWidget.show());
            });
      }
      function c(e) {
        var n = this,
          o = e.cfg,
          a = e.args,
          t = e.rspData.rsp.data || {};
        passport._load(M + L[S], !0, function () {
          C
            ? (C.setMakeOption && C.setMakeOption(t.authsid, t.bdstoken), C.show())
            : ((C = passport.pop.init({
                type: "accSetPwd",
                tangram: !0,
                color: n.config.color || "blue",
                apiOpt: {
                  u: n.config.u,
                  product: n.config.product,
                  authsid: t.authsid || "",
                  bdstoken: t.bdstoken || "",
                  staticPage: n.config.staticPage,
                },
                onHide: function () {
                  "1" != t.setpwdswitch && ((a.isCompleted = !0), o.onCancel && o.onCancel(a));
                },
                onSubmitSuccess: function () {
                  ((a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    t.hao123Param && U(t.hao123Param, n),
                    C.hide(),
                    o.onCancel && o.onCancel(a));
                },
              })),
              C.show());
        });
      }
      function l(e) {
        var n = this,
          o = e.cfg,
          a = e.args,
          t = e.rspData.rsp.data || {};
        passport._load(M + L[S], !0, function () {
          D
            ? (D.setMakeOption && D.setMakeOption(t.authsid, t.bdstoken), D.show())
            : ((D = passport.pop.init({
                type: "accSetPwd",
                tangram: !0,
                color: n.config.color || "blue",
                jumpset: "1" == t.jumpset ? 1 : 0,
                apiOpt: {
                  u: n.config.u,
                  product: n.config.product,
                  authsid: t.authsid || "",
                  bdstoken: t.bdstoken || "",
                  username: 1,
                  staticPage: n.config.staticPage,
                },
                onHide: function () {
                  "1" == t.jumpset && ((a.isCompleted = !0), o.onCancel && o.onCancel(a));
                },
                onSubmitSuccess: function () {
                  ((a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    t.hao123Param && U(t.hao123Param, n),
                    D.hide(),
                    o.onCancel && o.onCancel(a));
                },
              })),
              D.show());
        });
      }
      function p(e) {
        var n = this,
          o = e.cfg,
          a = e.args,
          t = e.rspData.rsp.data || {},
          r = "400413" === a.rsp.errInfo.no && "sms" === e.rspData.rsp.loginType ? "1" : "0";
        passport._load(M + L[S], !0, function () {
          T
            ? (T.reRender &&
                T.reRender({
                  lstr: t.lstr || "",
                  ltoken: t.ltoken || "",
                  token: t.authtoken || "",
                  loginproxy: t.loginproxy || "",
                  select: r,
                  loginType: e.rspData.rsp.loginType,
                }),
              T.show())
            : ((T = passport.pop.init({
                type: "secondCardVerify",
                tangram: !0,
                color: n.config.color || "blue",
                apiOpt: {
                  u: n.config.u,
                  product: n.config.product,
                  lstr: t.lstr || "",
                  ltoken: t.ltoken || "",
                  token: t.authtoken || "",
                  staticPage: n.config.staticPage,
                  select: r,
                  scscene: t.scscene || "",
                  scnewuser: t.scnewuser || "",
                  loginType: e.rspData.rsp.loginType,
                  loginproxy: t.loginproxy || "",
                },
                onloginSuccess: function () {
                  ((a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    t.hao123Param && U(t.hao123Param, n),
                    T.hide(),
                    o.onCancel && o.onCancel(a));
                },
                onSubmitSuccess: function (e, o) {
                  o.rsp.data.loginproxy &&
                    passport.data.jsonp(o.rsp.data.loginproxy).success(function (a) {
                      "0" === a.errInfo.no
                        ? (document.location.href =
                            "https://passport.baidu.com/v3/security/main/secondcardmodifyaccinfo?tpl=" +
                            n.config.product +
                            "&bdstoken=" +
                            o.rsp.data.bdstoken +
                            "&authsid=" +
                            o.rsp.data.mod_authsid +
                            "&u=" +
                            encodeURIComponent(a.data.u) +
                            "&loginType=" +
                            o.rsp.data.loginType +
                            "&hasUsername=" +
                            o.rsp.data.hasUsername)
                        : (e.target.getElement("secondCardError").innerHTML = n.lang.sysError);
                    });
                },
              })),
              T.show());
        });
      }
      function d(e) {
        var n = this,
          o = e.cfg,
          a = e.args,
          t = e.rspData.rsp.data || {};
        passport._load(M + L[S], !0, function () {
          P
            ? (P.reRender && P.reRender({ secauthsid: t.secauthsid, secbdstoken: t.secbdstoken }), P.show())
            : ((P = passport.pop.init({
                type: "secondCardList",
                tangram: !0,
                color: n.config.color || "blue",
                apiOpt: {
                  u: n.config.u,
                  product: n.config.product,
                  staticPage: n.config.staticPage,
                  secauthsid: t.secauthsid,
                  secbdstoken: t.secbdstoken,
                  regLink: n.constant.REG_URL,
                  source: n.config.source || "login",
                },
                onVerifySuccess: function (e) {
                  (P.hide(),
                    (a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    a && a.rsp && a.rsp.data && (a.rsp.data.exchange = e),
                    t.hao123Param && U(t.hao123Param, n),
                    o.onCancel && o.onCancel(a));
                },
                onRegSuccess: function () {
                  (P.hide(),
                    (a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    t.hao123Param && U(t.hao123Param, n),
                    o.onCancel && o.onCancel(a));
                },
                onHide: function () {
                  (n._logPass({}, { eventType: "pcapi-secondcardList-dialog-close" }),
                    n._ownerDialog && n._ownerDialog.show());
                },
              })),
              P.show());
        });
      }
      function u(e) {
        var n = this;
        ((e.rspData = e.rspData || {}),
          passport._load(M + L[x], !0, function () {
            k
              ? k.setToken(e.rspData.token, function () {
                  k.show();
                })
              : ((k = passport.pop.init({
                  type: "accConnect",
                  tangram: !0,
                  color: n.config.color || "",
                  apiOpt: {
                    u: n.config.u,
                    adtext: e.rspData.adtext,
                    product: n.config.product,
                    token: e.rspData.token,
                    staticPage: n.config.staticPage,
                  },
                  onConnectSuccess: function (e) {
                    var o = e.rsp;
                    ((o.connect = k), V(n, { rsp: o }), (e.returnValue = !1));
                  },
                })),
                k.show());
          }));
      }
      function g(e) {
        var n = this,
          o = e.cfg,
          a = e.args;
        ((e.rspData = e.rspData || {}),
          passport._load(M + L[S], !0, function () {
            y
              ? y.show()
              : ((y = passport.pop.init({
                  type: "accRealName",
                  tangram: !0,
                  color: n.config.color || "",
                  apiOpt: {
                    u: n.config.u,
                    product: n.config.product,
                    staticPage: n.config.staticPage,
                    action: "login",
                    realnameswitch: e.rspData.realnameswitch,
                    authsid: e.rspData.realnameauthsid,
                    ltoken: e.rspData.ltoken,
                    lstr: e.rspData.lstr,
                    cu: e.rspData.u,
                    overseas: n.config.overseas,
                  },
                  onHide: function () {
                    n.fire("RealHide", a);
                    "1" != e.rspData.realnameswitch && ((a.isCompleted = !0), o.onCancel && o.onCancel(a));
                  },
                  onverifyHide: function (t) {
                    n.realLoginHide ||
                      (e.rspData.loginproxy && t
                        ? passport.data.jsonp(e.rspData.loginproxy + "&authsid=" + t).success(function (e) {
                            0 == e.errInfo.no
                              ? ((a.isCompleted = !0),
                                o.onCompleted &&
                                  o.onCompleted(e, function () {
                                    o.onCancel && o.onCancel(a);
                                  }))
                              : (n._ownerDialog && n._ownerDialog.show(),
                                (n.getElement("error").innerHTML = n.lang.sysError));
                          })
                        : ((a.isCompleted = !0),
                          e.rspData.hao123Param && "1" == e.rspData.realnameswitch && U(e.rspData.hao123Param, n),
                          o.onCancel && o.onCancel(a)));
                  },
                  onSubmitSuccess: function (t, r) {
                    e.rspData.loginproxy && r.rsp.data.authsid
                      ? passport.data
                          .jsonp(e.rspData.loginproxy + "&authsid=" + r.rsp.data.authsid)
                          .success(function (e) {
                            if (0 === +e.errInfo.no) {
                              if (
                                ((a.isCompleted = !0),
                                (n.realLoginHide = !0),
                                y.hide(),
                                1 === +e.data.guideUpgradeMobile)
                              )
                                return void m.call(n, e, a);
                              o.onCompleted &&
                                o.onCompleted(e, function () {
                                  o.onCancel && o.onCancel(a);
                                });
                            } else
                              (y.hide(),
                                n._ownerDialog && n._ownerDialog.show(),
                                (n.getElement("error").innerHTML = n.lang.sysError));
                          })
                      : ((a.isCompleted = !0),
                        a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                        e.rspData.hao123Param && e.rspData.realnameswitch + "" == "1" && U(e.rspData.hao123Param, n),
                        (n.realLoginHide = !0),
                        y.hide(),
                        o.onCancel && o.onCancel(a));
                  },
                })),
                y.show());
          }));
      }
      function m(e, n) {
        var o = this;
        passport._load(M + L[S], !0, function () {
          var a = passport.pop.init({
            type: "bindGuide",
            tangram: !0,
            color: o.config.color || "",
            apiOpt: {
              u: o.config.u,
              product: o.config.product,
              staticPage: o.config.staticPage,
              action: "login",
              realnameswitch: e.data.realnameswitch,
              authsid: e.data.realnameauthsid,
              ltoken: e.data.ltoken,
              upgradeMobileToken: e.data.upgradeMobileToken,
              bindMobileToken: e.data.bind_mobile_token,
              lstr: e.data.lstr,
              cu: e.data.u,
              overseas: o.config.overseas,
              tip: W.bindTip,
              successTip: (e && e.data && e.data.upgradeMobile) + W.verifySuccess,
            },
            onSubmitSuccess: function () {
              (a.hide(), V(o, n));
            },
            onHide: function () {
              V(o, n);
            },
          });
          a.show();
        });
      }
      function f(e) {
        var n = this,
          o = (e.cfg, e.args);
        ((e.rspData = e.rspData || {}),
          passport._load(M + L[S], !0, function () {
            E
              ? E.show()
              : ((E = passport.pop.init({
                  type: "bindGuide",
                  tangram: !0,
                  color: n.config.color || "",
                  apiOpt: {
                    type: "verify",
                    u: n.config.u,
                    product: n.config.product,
                    staticPage: n.config.staticPage,
                    action: "login",
                    realnameswitch: e.rspData.realnameswitch,
                    authsid: e.rspData.realnameauthsid,
                    ltoken: e.rspData.ltoken,
                    upgradeMobileToken: e.rspData.upgrade_mobile_token,
                    bindMobileToken: e.rspData.bind_mobile_token,
                    lstr: e.rspData.lstr,
                    cu: e.rspData.u,
                    overseas: n.config.overseas,
                    tip: W.checkUnVPhone + (e.rspData.upgrade_mobile || "") + "，" + W.setBindPhone,
                    successTip: W.loginsuccess,
                  },
                  onSubmitSuccess: function () {
                    (E.hide(), V(n, o));
                  },
                  onHide: function () {
                    V(n, o);
                  },
                })),
                E.show());
          }));
      }
      function h(e) {
        var n = this,
          o = e.cfg,
          a = e.args;
        ((e.rspData = e.rspData || {}),
          passport._load(M + L[I], !0, function () {
            var t = passport.pop.ArmorWidget("bindmobile", {
              token: e.rspData.bindMobileToken || "",
              title: W.bindPhone,
              msg: "",
              action: "bindmobile",
              overseas: 1,
              entryScene: "login",
              lstr: e.rspData.lstr || "",
              ltoken: e.rspData.ltoken || "",
              tpl: e.rspData.tpl || "",
              u: e.rspData.u || "",
              fromPv: "upgradeMobile",
              switchToBind: e.rspData.switchtobind || 0,
              authsid: e.rspData.bindwidgetauthsid || "",
              authid: e.rspData.authid || "",
              hideEmailAndPhoneText: n.config.hideEmailAndPhoneText || !1,
              auth_title: W.bindPhone,
              auth_msg: W.enterBind,
              staticPage: n.config.staticPage || "",
              onSubmitSuccess: function () {
                e.rspData.loginproxy
                  ? passport.data.jsonp(e.rspData.loginproxy).success(function (e) {
                      0 === +e.errInfo.no
                        ? ((a.isCompleted = !0), V(n, a))
                        : (n._ownerDialog && n._ownerDialog.show(),
                          (n.getElement("error").innerHTML = n.lang.sysError));
                    })
                  : ((a.isCompleted = !0),
                    a && a.rsp && a.rsp.errInfo && (a.rsp.errInfo.no = "0"),
                    e.rspData.hao123Param && e.rspData.realnameswitch + "" == "1" && U(e.rspData.hao123Param, n),
                    o.onCancel && o.onCancel(a));
              },
            });
            t.show();
          }));
      }
      function b(e, n) {
        try {
          var o = n.data && n.data.bySuccessUParams;
          if (o && "object" == typeof o) {
            var a = [];
            for (var t in o)
              o.hasOwnProperty(t) &&
                void 0 !== o[t] &&
                null !== o[t] &&
                a.push(encodeURIComponent(t) + "=" + encodeURIComponent(o[t]));
            if (a.length > 0) {
              var r = e.rsp.data.u || "";
              e.rsp.data.u = r + (r.indexOf("?") > -1 ? "&" : "?") + a.join("&");
            }
          }
        } catch (i) {
          console.error("bySuccessUParams error:", i);
        }
      }
      function w(e, a, i) {
        function m(n) {
          var o = new Image();
          ((o.onload = o.onerror =
            function () {
              ((o.onload = o.onerror = null), (o = null));
            }),
            (o.src = M + "/img/v.gif?type=" + n + "&tt=" + new Date().getTime()),
            (a.isCompleted = !0),
            w.hao123Param && U(w.hao123Param, e),
            w && w.sync_id && _(w.sync_id, E),
            E.confirmWidgetMobileSure.hide(),
            i.onCancel && i.onCancel(a));
        }
        var b = a.rsp.errInfo.no,
          w = a.rsp.data,
          y = a.rsp && a.rsp.errInfo && a.rsp.errInfo.msg;
        i = i || {};
        var C = w && ("risk" === w.secstate || "cheat" === w.secstate);
        if (w && w.connectType && e.config.connect) {
          var D = e.fire("beforeWarning", { args: a });
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            v(M + "/passApi/css/uni_accConnect_ab6dda9.css"),
            u.apply(e, [{ args: a, rspData: w, cfg: i, errno: b }]),
            !1
          );
        }
        if (18 === +b)
          return (
            e._logNaMonitor("pc_login_real_name_action"),
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            g.apply(e, [{ args: a, rspData: w, cfg: i, errno: b }]),
            w.hao123Param && w.realnameswitch + "" != "1" && U(w.hao123Param, e),
            !1
          );
        if (25 === +b) {
          (e._logNaMonitor("pc_bind_action"),
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"));
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return w.authid
            ? void t.apply(e, [{ args: a, rspData: w, cfg: i, errno: C && 120021 !== +b ? "riskCheat" : b }])
            : (h.apply(e, [{ args: a, rspData: w, cfg: i, errno: b }]),
              w.hao123Param && w.realnameswitch + "" != "1" && U(w.hao123Param, e),
              !1);
        }
        if (+w && w.guide_upgrade_mobile && 1 === w.guide_upgrade_mobile)
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            e._logNaMonitor("pc_login_confirm_mobile"),
            f.call(e, { args: a, rspData: w, cfg: i, errno: b }),
            w.hao123Param && w.realnameswitch + "" != "1" && U(w.hao123Param, e),
            !1
          );
        if (20 === +b) {
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e.getElement("smsError") && (e.getElement("smsError").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            c.apply(e, [{ args: a, rspData: a, cfg: i }]),
            w.hao123Param && "1" != w.setpwdswitch && U(w.hao123Param, e),
            !1
          );
        }
        if (22 === +b) {
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e.getElement("smsError") && (e.getElement("smsError").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            l.apply(e, [{ args: a, rspData: a, cfg: i }]),
            w.hao123Param && U(w.hao123Param, e),
            !1
          );
        }
        if ("400413" === b || "400414" === b) {
          e._logNaMonitor("pc_login_second_card");
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e.getElement("smsError") && (e.getElement("smsError").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            p.apply(e, [{ args: a, rspData: a, cfg: i }]),
            w.hao123Param && U(w.hao123Param, e),
            !1
          );
        }
        if (400425 === +b) {
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          if (
            (e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e.getElement("smsError") && (e.getElement("smsError").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            w.authid)
          )
            return void t.apply(e, [{ args: a, rspData: w, cfg: i, errno: b }]);
        }
        if ("400704" === b || "400706" === b) {
          e._logNaMonitor("pc_login_second_card");
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e.getElement("smsError") && (e.getElement("smsError").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            d.apply(e, [{ args: a, rspData: a, cfg: i }]),
            w.hao123Param && U(w.hao123Param, e),
            !1
          );
        }
        if (19 == b) {
          var E = this,
            w = a.rsp.data,
            D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            null == E.confirmWidgetLabel
              ? passport._load(M + "/passApi/js/uni_wrapper.js", !0, function () {
                  ((E.confirmWidgetLabel = passport.pop.init({
                    type: "confirmWidget",
                    tangram: !0,
                    width: 430,
                    height: 150,
                    apiOpt: {
                      contentHTML:
                        '                                        <div class="pass-confirmContent-wrapper-Label">                                            <div class="pass-confirmContent-msg-Label">                                                <span class="img-class" ></span>                                            </div>                                            <ul class="pass-confirmContent-descmsg-Label" style="margin-left:20px;">                                                <p>' +
                        W.confirmWidgetTip1 +
                        "</p>                                                <p>" +
                        W.confirmWidgetTip2 +
                        "</p>                                                <p>" +
                        W.confirmWidgetTip2 +
                        "</p>                                            </ul>                                        </div>",
                    },
                    onConfirmClose: function () {
                      (E.confirmWidgetLabel.hide(), V(e, a));
                    },
                    onRender: function () {
                      ((E.confirmWidgetLabel.getElement("confirm_cancel").style.display = "none"),
                        (E.confirmWidgetLabel.getElement("confirm_continue").style.display = "none"),
                        (E.confirmWidgetLabel.getElement("confirmWidget_footer").style.display = "none"));
                    },
                  })),
                    E.confirmWidgetLabel.show());
                })
              : E.confirmWidgetLabel.show(),
            w.hao123Param && U(w.hao123Param, e),
            w && w.sync_id && _(w.sync_id, E),
            !1
          );
        }
        if (23 == b) {
          var E = this,
            w = a.rsp.data,
            D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            null == E.confirmWidgetMobileSure
              ? passport._load(M + "/passApi/js/uni_wrapper.js", !0, function () {
                  ((E.confirmWidgetMobileSure = passport.pop.init({
                    type: "confirmWidget",
                    tangram: !0,
                    titleText: "非常重要",
                    width: 430,
                    apiOpt: {
                      Cancel: "不需要解绑",
                      Continue: "申诉去解绑",
                      contentHTML:
                        '<div class="pass-confirmContent-wrapper-sureConfirm"><div class="pass-confirmContent-wrapper-msg"><p><span class="pass-confirmContent-redcolor" id="pass-mobile-sure-num">' +
                        w.phoneNumber +
                        '</span>是您绑定的手机号,请确认该手机号是否还在使用,为了账号安全请及时解绑不使用的手机。</p></div><div class="pass-confirmwidget-bottom"><span id="pass-mobile-sure-btn" class="pass-button pass-button-grey cancel">不需要解绑</span><a href="' +
                        w.appealurl +
                        '" id="pass-appeal-btn" class="pass-button pass-button-grey continue" target="new">申诉去解绑</a></div></div>',
                    },
                    onRender: function () {
                      ((E.confirmWidgetMobileSure.getElement("confirm_cancel").style.display = "none"),
                        (E.confirmWidgetMobileSure.getElement("confirm_continue").style.display = "none"),
                        (E.confirmWidgetMobileSure.getElement("confirmWidget_footer").style.display = "none"),
                        baidu(document.getElementById("pass-mobile-sure-btn")).on("click", function () {
                          m("mobileSurePC");
                        }),
                        baidu(document.getElementById("pass-appeal-btn")).on("click", function () {
                          m("appealMobilePC");
                        }));
                    },
                    onConfirmClose: function () {
                      m("mobileSureClosePC");
                    },
                    onConfirmContinue: function () {},
                    onConfirmCancel: function () {},
                  })),
                    E.confirmWidgetMobileSure.show());
                })
              : (document.getElementById("pass-mobile-sure-num") &&
                  (document.getElementById("pass-mobile-sure-num").html = w.phoneNumber || ""),
                E.confirmWidgetMobileSure.show()),
            w.hao123Param && U(w.hao123Param, e),
            w && w.sync_id && _(w.sync_id, E),
            !1
          );
        }
        if (401007 == b) {
          e._logNaMonitor("pc_login_multi_choice");
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            r.apply(e, [{ rspData: w }]),
            !1
          );
        }
        if (120016 == b || 400032 == b || 400034 == b) {
          e._logNaMonitor("pc_bind_action");
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            n.apply(e, [{ errno: b, args: a, cfg: i }]),
            w.hao123Param && U(w.hao123Param, e),
            !1
          );
        }
        if (5 == b || 400031 == b || 120019 == b || 120021 == b || 400023 == b || C) {
          e._logNaMonitor("pc_login_verify_transfer");
          var D = e.fire("beforeWarning", { args: a });
          if (!D) return;
          return (
            e.getElement("error") && (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            w.authid
              ? void t.apply(e, [{ args: a, rspData: w, cfg: i, errno: C && 120021 != b ? "riskCheat" : b }])
              : (v(M + "/passApi/css/uni_forceverify_064a5ef.css"),
                o.apply(e, [{ args: a, rspData: w, cfg: i, errno: C && 120021 != b ? "riskCheat" : b }]),
                !1)
          );
        }
        if (400037 == b) {
          var D = e.fire("beforeWarning", a);
          if (!D) return;
          return (
            (e.getElement("error").innerHTML = ""),
            e._ownerDialog && e._ownerDialog.hide("unHide"),
            s.apply(e, [{ args: a, rspData: a, cfg: i }]),
            w.hao123Param && U(w.hao123Param, e),
            !1
          );
        }
        if (400041 === +b) {
          if (O) return void (e.getElement("error").innerHTML = y);
          document.location.href = "https://passport.baidu.com/?getpassappeal&tpl=" + e.config.product;
        }
        return !0;
      }
      function v(e) {
        if (!j[e]) {
          j[e] = !0;
          var n = document.createElement("link");
          ((n.rel = "stylesheet"),
            (n.type = "text/css"),
            (n.href = e),
            document.getElementsByTagName("head")[0].appendChild(n));
        }
        return !0;
      }
      function _(e, n) {
        var o =
          (document.location.protocol.toLowerCase(),
          "https://passport.baidu-int.com/v3/api/login/syncintbduss?tpl=" +
            n.config.product +
            "&sync_id=" +
            e +
            "&t=" +
            new Date().getTime());
        passport.data.request.load(o);
      }
      var y,
        C,
        D,
        E,
        k,
        T,
        P,
        M = window.location
          ? "http:" === window.location.protocol.toLowerCase()
            ? "http://passport.baidu.com"
            : "https://passport.baidu.com"
          : "http:" === document.location.protocol.toLowerCase()
            ? "http://passport.baidu.com"
            : "https://passport.baidu.com",
        L = {
          uni_armorwidget: "/passApi/js/uni_armorwidget_c71aab9.js",
          uni_forceverify: "/passApi/js/uni_forceverify_bc920d3.js",
          uni_accConnect: "/passApi/js/uni_accConnect_wrapper.js",
          uni_wrapper: "/passApi/js/uni_wrapper.js",
        },
        S = "uni_wrapper",
        H = "uni_forceverify",
        I = "uni_armorwidget",
        x = "uni_accConnect",
        j = {},
        W = passport.err.getCurrent().labelText.login,
        A = passport.err.getCurrent().hooksText,
        R = passport.passLang,
        N = passport && passport.enTpls,
        B = passport.tpl,
        O = N && N.includes(B),
        V = function (e, n) {
          function o() {
            var o = "pc_" + e.currentLoginType + "_loginv4_success";
            e._logNaMonitor(o);
            var a =
              ({ tpl: e.config.product || "", page: e.config.page || "", traceId: e.urlData.traceId || "" },
              { eventType: "pc_login_success" });
            (e._logPass(e.urlData, a),
              e._logPass(e.urlData, { eventType: "pc-" + e.config.loginType + "-submit-success" }));
            var t = e.fire("loginSuccess", n);
            t &&
              (window.YY_TPL_CONFIG.indexOf(e.config.product + ",") > -1
                ? passport.data
                    .getAccessToken({
                      client: "",
                      clientfrom: "",
                      tpl: e.config.product,
                      client_id: "Bxr73Efet8HjR5Tr0HqcgDKr",
                    })
                    .success(function (o) {
                      var a = o.data,
                        t = e.fireEvent("yyLoginSuccess", { rsp: a });
                      t &&
                        (window.location
                          ? (window.location.href =
                              n.rsp.data.u + "?accesstoken=" + a.access_token + "&expire=" + a.expires_in)
                          : (document.location.href =
                              n.rsp.data.u + "?accesstoken=" + a.access_token + "&expire=" + a.expires_in));
                    })
                : window.location
                  ? (window.location.href = n.rsp.data.u)
                  : (document.location.href = n.rsp.data.u));
          }
          try {
            if (n.rsp && n.rsp.data) {
              var a = n.rsp.data.RefreshToken || (n.rsp.data.session && n.rsp.data.session.RefreshToken);
              if (!a) return void o();
              passport._load(
                location.protocol + "//wappass.baidu.com/static/waplib/passRefreshTk.js?tt=" + new Date().getTime(),
                !0,
                function () {
                  var e = (window.passRefreshTk && window.passRefreshTk.encryptStoreToken) || "";
                  (!e && o(),
                    a &&
                      e &&
                      e(
                        a,
                        function () {
                          o();
                        },
                        function () {
                          o();
                        },
                        !0,
                      ));
                },
              );
            }
          } catch (t) {
            o();
          }
        },
        q = function () {
          var e = 0,
            n = 0,
            o = null,
            a = /msie (\d+\.\d+)/i.test(navigator.userAgent);
          if ((a || console.log(navigator.plugins["Shockwave Flash"]), a)) {
            var t = new ActiveXObject("ShockwaveFlash.ShockwaveFlash");
            t && ((e = 1), (n = t.GetVariable("$version")));
          } else if (
            navigator &&
            navigator.plugins &&
            navigator.plugins.length > 0 &&
            navigator.plugins["Shockwave Flash"]
          ) {
            var t = navigator.plugins["Shockwave Flash"];
            t && ((o = t.filename), (e = 1), (n = t.version ? t.version : t.description));
          }
          return { isexists: e, ver: n, filename: o };
        },
        U = function (e, n) {
          if (!((n && n.config && n.config.noSynBdu && 1 === n.config.noSynBdu) || O)) {
            var o = document.location.protocol.toLowerCase(),
              a = encodeURIComponent(e);
            if ("copyright" === n.config.product) {
              var t = "//newcopyright.baidu-int.com/login/crossdomain?bdu=" + a;
              passport.data.request.load(o + t + "&t=" + new Date().getTime());
            }
            if ("eventgraph" === n.config.product) {
              var r = "//eventgraph.baidu-int.com/event_api/api/get_access_bduss?bdu=" + a;
              passport.data.request.load(o + r + "&t=" + new Date().getTime());
            }
            if ("roadugc" === n.config.product) {
              var i = "//inner-lutao.baidu-int.com/audit/api/crossdomain?bdu=" + a;
              passport.data.request.load(o + i + "&t=" + new Date().getTime());
            }
            if ("map_bus_data" === n.config.product) {
              var s = "//busjiucuo.baidu-int.com/subwaybus/passport/crossdomain?bdu=" + a,
                c = "//prebusjiucuo.baidu-int.com/subwaybus/passport/crossdomain?bdu=" + a;
              (passport.data.request.load(o + s + "&t=" + new Date().getTime()),
                passport.data.request.load(o + c + "&t=" + new Date().getTime()));
            }
            if ("zhengxin" === n.config.product) {
              var l = "//www.aiqicha.com/login/loginAjax?bdu=" + a,
                p = "//www.aiqicha.cn/login/loginAjax?bdu=" + a;
              (passport.data.request.load(o + l + "&t=" + new Date().getTime()),
                passport.data.request.load(o + p + "&t=" + new Date().getTime()));
            }
            if ("health_pc_cer" === n.config.product) {
              var d = "//www.baidujk.com/harbor/pc/crossdomain?bdu=" + a;
              passport.data.request.load(o + d + "&t=" + new Date().getTime());
            }
            if ("netdisk_sandbox" === n.config.product) {
              var i = "//sandbox.pan.baidu-int.com/pass/crossdomain?bdu=" + a;
              passport.data.request.load(o + i + "&t=" + new Date().getTime());
            }
            if ("dipr" === n.config.product) {
              var i = "//www.bddipr.com/cms/crossdomain/login?bdu=" + a;
              passport.data.request.load(o + i + "&t=" + new Date().getTime());
            }
            ("wantalk" === n.config.product &&
              passport.data.request.load(
                o + "//wantalk.com/api/wanhua/login/crossdomain/pc?bdu=" + a + "&t=" + new Date().getTime(),
              ),
              "zhengxin" === n.config.product &&
                passport.data.request.load(
                  o + "//faxingbao.com/api/crossdomain?bdu=" + a + "&t=" + new Date().getTime(),
                ),
              passport.data.request.load(
                o + "//developer.apollo.auto/apollo/sync_bduss?bdu=" + a + "&t=" + new Date().getTime(),
              ));
            var u = "//baike.baidu.hk/common/api/crossdomain?bdu=" + a;
            passport.data.request.load(o + u + "&t=" + new Date().getTime());
            var g = "//www.yoojia.com/sync/crossdomain?bdu=" + a;
            return (
              passport.data.request.load(o + g + "&t=" + new Date().getTime()),
              passport.data.request.load(
                o + "//user.hao123.com/static/crossdomain.php?bdu=" + a + "&t=" + new Date().getTime(),
              )
            );
          }
        };
      e.login = {
        loginSuccess: function (e, n) {
          function o() {
            var o = w(e, n, {
              onCancel: function () {
                V(e, n);
              },
            });
            o && V(e, n);
          }
          var a = /msie (\d+\.\d+)/i.test(navigator.userAgent);
          if (
            (e.config.setWebToClient && !a && i.apply(e, [{ args: n }]),
            (e.config.noSynBdu && 1 === e.config.noSynBdu) || O)
          ) {
            var t = w(e, n, {
              onCancel: function () {
                V(e, n);
              },
            });
            t && V(e, n);
          } else
            n.rsp && n.rsp.data && n.rsp.data.hao123Param
              ? U(n.rsp.data.hao123Param, e).success(function () {
                  o();
                })
              : o();
          return (
            n.rsp.data && n.rsp.data.sync_id && _(n.rsp.data.sync_id, e),
            { preventEvent: !0, preventDefault: !0 }
          );
        },
        loginError: function (e, n) {
          return (
            w(e, n, {
              onCompleted: function (n, o) {
                (e.config.noSynBdu && 1 === e.config.noSynBdu) || O ? o() : U(n.data.hao123Param, e).success(o);
              },
              onCancel: function (n) {
                n && n.isCompleted && V(e, { rsp: n.rsp });
              },
            }),
            { preventEvent: !1, preventDefault: !1 }
          );
        },
        connectNeedBind: function (e, n) {
          return (
            w(e, n, { onCompleted: function () {}, onCancel: function () {} }),
            { preventEvent: !1, preventDefault: !1 }
          );
        },
        validateAllError: function (e, n) {
          var o = n.validates ? e.getElement(n.validates[n.validates.length - 1].field) : "";
          return (o && o.focus && o.focus(), { preventEvent: !1, preventDefault: !1 });
        },
      };
    })(passport.hook));
  var validatorType = window.passport && "en" === window.passport.passLang ? "en" : "default";
  magic.passport = baidu.lang
    .createClass(
      function () {
        this._validateInfo = {};
      },
      { type: "magic.passport", superClass: magic.Base },
    )
    .extend({
      _getRegularField: function (e) {
        var t = e.pwd ? "password" : "text",
          a = this,
          n = 'autocomplete="off"',
          i = e.maxLength ? 'maxlength="' + e.maxLength + '"' : "",
          r = e.tip ? e.tip : "",
          o = e.value ? e.value : "",
          l = e.field + "" == "verifycode" ? "none" : "",
          s = "";
        s =
          "text" === t ? '<input type="text" style="display:none;">' : '<input type="password" style="display: none;">';
        var c =
          '<p id="' +
          a.$getId(e.field + "Wrapper") +
          '" class="pass-form-item pass-form-item-' +
          e.field +
          ("setpwd" === e.type ? " pass-new-set-pwd" : "") +
          '" style="display:' +
          l +
          '">' +
          (e.label
            ? '<label for="' +
              a.$getId(e.field) +
              '" id="' +
              a.$getId(e.field + "Label") +
              '" class="pass-label pass-label-' +
              e.field +
              '">' +
              e.label +
              "</label>"
            : "") +
          s +
          '<input id="' +
          a.$getId(e.field) +
          '" type="' +
          t +
          '" name="' +
          e.field +
          '" class="pass-text-input pass-text-input-' +
          e.field +
          '" ' +
          i +
          (e.placeholder ? 'placeholder="' + e.placeholder + '" ' : "") +
          n +
          (e.disabled ? '" disabled' : "") +
          ' value="' +
          o +
          '"/>' +
          (e.noError
            ? ""
            : '<span id="' +
              a.$getId(e.field + "Error") +
              '" class="pass-item-error pass-item-error-' +
              e.field +
              '"></span>') +
          (e.hasSucc
            ? '<span id="' +
              a.$getId(e.field + "Succ") +
              '" class="pass-item-succ pass-item-succ-' +
              e.field +
              '" style="display:none;"></span>'
            : "") +
          '<span id="' +
          a.$getId(e.field + "Tip") +
          '" class="pass-item-tip pass-item-tip-' +
          e.field +
          '" style="display:none"><span id="' +
          a.$getId(e.field + "TipText") +
          '" class="pass-item-tiptext pass-item-tiptext-' +
          e.field +
          '">' +
          r +
          "</span></span></p>";
        return c;
      },
      _getHiddenField: function (e, t) {
        var a,
          n = this,
          i = '<p id="' + n.$getId(t || "hiddenFields") + '" style="display:none">';
        for (var r in e)
          e.hasOwnProperty(r) &&
            ((a =
              "string" == typeof e[r]
                ? e[r]
                    .replace(/</g, "&lt;")
                    .replace(/>/g, "&gt;")
                    .replace(/"/g, "&quot;")
                    .replace(/\x22/g, "&quot;")
                    .replace(/\x27/g, "&#39;")
                : e[r]),
            (i +=
              '<input type="hidden" id="' +
              n.$getId((t ? t + "_" : "") + r) +
              '" name="' +
              r +
              '" value="' +
              a +
              '" />'));
        return (i += "</p>");
      },
      _setEvent: function () {
        var e = this,
          t = this.getElement(),
          a = function (t) {
            e._eventHandler.entrance.apply(e, [t]);
          };
        (baidu(e.getElement()).on("resize", function () {
          var e = (navigator.userAgent, !navigator.userAgent.match(/OS [1-8]_\d[_\d]* like Mac OS X/i)),
            t = !!navigator.userAgent.toString().match(/\(i[^;]+;( U;)? CPU.+Mac OS X/),
            a = navigator.userAgent.toString().indexOf("iPad");
          if (e && t && null != a) {
            var n = document.getElementsByClassName("popBox");
            null != n &&
              n.length > 0 &&
              (n[0].style.height =
                window.screen.height > document.body.clientHeight
                  ? window.screen.height * (window.screen.height / document.body.clientHeight) + 120 + "px"
                  : window.screen.height * (window.screen.height / document.body.clientHeight));
          }
        }),
          baidu(e.getElement("form")).on("submit", a),
          baidu(e.getElement("license")).on("click", a),
          baidu(e.getElement("verifyCodeChange")).on("click", a),
          baidu(e.getElement("verifyCodeSend")).on("click", a),
          baidu(e.getElement("smsVcodeSend")).on("click", a),
          baidu(t).delegate(".pass-suggest-item label", "click", a),
          baidu(".pass-text-input", e.getElement()).on({
            focus: a,
            blur: a,
            change: a,
            keyup: a,
            mouseover: a,
            mouseout: a,
          }));
      },
      _validator: {
        confStorage: {},
        builtInMsg: {
          default: {
            required: "请您输入%s",
            phone: "手机号码格式不正确",
            email: "邮箱格式不正确",
            idcard: "身份证格式不正确",
            internetOfThingsRule: "请勿使用物联网手机号",
          },
          en: {
            required: "Please enter %s",
            phone: "Incorrectly formatted cell phone number",
            email: "Incorrect mailbox format",
            idcard: "Incorrectly formatted ID card",
            internetOfThingsRule: "Do not use IoT cell phone numbers",
          },
        },
        builtInRules: {
          default: {
            required: /\S+/,
            phone: /^1[3456789]\d{9}$/,
            email: /^([a-zA-Z0-9_\.\-\+])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$/,
            idcard: /(^\d{15}$)|(^\d{17}(\d|X|x)$)/,
            internetOfThingsRule: function (e) {
              return 13 === e.value.length
                ? !/^(14[01234])\d*$/.test(e.value)
                : !/^((14[012345689])|(1740)|(1741[12]))\d*$/.test(e.value);
            },
          },
          en: {
            required: /\S+/,
            phone: /^\d+$/,
            email: /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/,
            idcard: /(^\d{15}$)|(^\d{17}(\d|X|x)$)/,
            internetOfThingsRule: function (e) {
              return 13 === e.value.length
                ? !/^(14[01234])\d*$/.test(e.value)
                : !/^((14[012345689])|(1740)|(1741[12]))\d*$/.test(e.value);
            },
          },
        },
        init: function (e, t) {
          this.confStorage[e.$getId()] = t;
        },
        validate: function (e, t, a, n) {
          var i = e.getElement(t),
            r = { field: t },
            o = /^\s*(.+?)\s*$/;
          if (!i || 0 === +i.offsetHeight) return !1;
          for (
            var l = i.value.replace(o, "$1"), s = this.confStorage[e.$getId()][t], c = s.rules, d = 0, u = c.length;
            u > d;
            d++
          ) {
            var p,
              f = c[d],
              g = this.builtInRules[validatorType][f];
            if (
              ((p =
                "[object function]" === Object.prototype.toString.call(g).toLowerCase()
                  ? g.call(e, i, n)
                  : new RegExp(g).test(l)),
              !p)
            )
              return (
                (r.error = !0),
                (e._validateInfo[t] = 0),
                (r.msg = this.builtInMsg[validatorType][f].replace(/\%s/, s.desc)),
                e._validateError(r),
                void a.error(r)
              );
          }
          s.asyncRule
            ? s.asyncRule.call(
                e,
                {
                  success: function (n) {
                    ((r.error = !1), (r.data = n.data), (e._validateInfo[t] = 1), e._validateSuccess(r), a.success(r));
                  },
                  error: function (n) {
                    ((r.error = !0), (e._validateInfo[t] = 0), (r.msg = n.msg), e._validateError(r), a.error(r));
                  },
                },
                n,
              )
            : ((r.error = !1), (e._validateInfo[t] = 1), e._validateSuccess(r), a.success(r));
        },
        validateAll: function (e, t, a) {
          function n() {
            ((l = !0), o ? t && t.error(s) : t && t.success(s));
          }
          var i = this.confStorage[e.$getId()],
            r = 0,
            o = !1,
            l = !1,
            s = [],
            c = this._getActiveValidate(e, !0);
          for (var d in i) {
            if (l) break;
            this.validate(
              e,
              d,
              {
                success: function (e) {
                  (r++, s.push(e), r === c && n());
                },
                error: function (e) {
                  return ((o = !0), r++, s.push(e), a ? void n() : void (r === c && n()));
                },
              },
              "all",
            );
          }
        },
        _getActiveValidate: function (e, t) {
          var a = this.confStorage[e.$getId()],
            n = [];
          for (var i in a) {
            var r = e.getElement(i);
            r && 0 !== r.offsetHeight && n.push(r);
          }
          return t ? n.length : n;
        },
        addRule: function (e, t) {
          var a = {};
          ((a[e] = t), baidu.extend(this.builtInRules[validatorType], a));
        },
        addMsg: function (e, t) {
          var a = {};
          ((a[e] = t), baidu.extend(this.builtInMsg[validatorType], a));
        },
      },
      validate: function (e, t) {
        var a = this,
          n = a.fireEvent("beforeValidate", { validate: { field: e, ele: baidu(a.getElement(e)) } });
        n &&
          a._validator.validate(a, e, {
            success: function (e) {
              var n = a.fireEvent("validateSuccess", { validate: e });
              n && t && t.success && t.success(e);
            },
            error: function (e) {
              var n = a.fireEvent("validateError", { validate: e });
              n && t && t.error && t.error(e);
            },
          });
      },
      validateAll: function (e, t) {
        var a = this;
        t = "boolean" == typeof e ? e : t ? t : !1;
        var n = a.fireEvent("beforeValidateAll");
        n &&
          a._validator.validateAll(
            a,
            {
              success: function (t) {
                var n = a.fireEvent("validateAllSuccess", { validates: t });
                n && e && e.success && e.success(t);
              },
              error: function (t) {
                var n = a.fireEvent("validateAllError", { validates: t });
                n && e && e.error && e.error(t);
              },
            },
            t,
          );
      },
      getValidateStatus: function (e) {
        return this._validateInfo.hasOwnProperty(e) ? this._validateInfo[e] : -1;
      },
      setValidateSuccess: function (e) {
        var t = this;
        ((t._validateInfo[e] = 1), t._validateSuccess({ field: e }));
      },
      setValidateError: function (e, t, a) {
        var n = this;
        ((n._validateInfo[e] = 0), n._validateError({ field: e, msg: t }, a));
      },
      setGeneralError: function (e) {
        this.getElement("error") && (this.getElement("error").innerHTML = e);
      },
      clearGeneralError: function () {
        this.getElement("error") && (this.getElement("error").innerHTML = "");
      },
      _isSupportPlaceholder: function () {
        return "placeholder" in document.createElement("input");
      },
      _getPlaceholder: function (e) {
        for (var t = this, a = {}, n = "", i = {}, r = 0; r < e.length; r++) {
          if (
            ((n = t.lang[e[r].placeholder]),
            0 !== +e[r].clearbtn &&
              ((i[r] = baidu(
                '<span id="' +
                  t.$getId(e[r].label + "_clearbtn") +
                  '" class="pass-clearbtn pass-clearbtn-' +
                  e[r].label +
                  '" style="display:none;"></span>',
              )),
              baidu(t.getElement(e[r].label)).after(i[r])),
            t._isSupportPlaceholder() && !baidu.browser.ie)
          )
            try {
              baidu(t.getElement(e[r].label)).attr({ placeholder: n });
            } catch (o) {
              var l = t.getElement(e[r].label);
              l.setAttribute("placeholder", n);
            }
          else
            ((a[r] = baidu(
              '<span id="' +
                t.$getId(e[r].label + "_placeholder") +
                '" class="pass-placeholder pass-placeholder-' +
                e[r].label +
                '">' +
                n +
                "</span>",
            )),
              baidu(t.getElement(e[r].label)).after(a[r]));
          t._rendEventPlaceholder(t.getElement(e[r].label), a[r], i[r]);
        }
      },
      _getCookie: function (e) {
        var t,
          a = new RegExp("(^| )" + e + "=([^;]*)(;|$)");
        return (t = document.cookie.match(a)) ? unescape(t[2]) : null;
      },
      _setCookie: function (e, t, a) {
        try {
          (a && ((a.path = a.path || "/"), (a.domain = a.domain || ".iqiyi.com")), (a = a || {}));
          var n = a.expires;
          ("number" == typeof a.expires && ((n = new Date()), n.setTime(n.getTime() + a.expires)),
            (document.cookie =
              e +
              "=" +
              t +
              (a.path ? "; path=" + a.path : "") +
              (n ? "; expires=" + n.toGMTString() : "") +
              (a.domain ? "; domain=" + a.domain : "") +
              (a.secure ? "; secure" : "")));
        } catch (i) {}
      },
      _logNaMonitor: function (e) {
        var t = { page: "loginv4", tpl: this.config.product || "", module: "wapna", monitorType: e },
          a = { eventType: "na-monitor" };
        this._logPass(t, a, 1023);
      },
      _logPass: function (e, t, a) {
        function n(e) {
          var t,
            a,
            n,
            i,
            r,
            o,
            l = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
          for (n = e.length, a = 0, t = ""; n > a;) {
            if (((i = 255 & e.charCodeAt(a++)), a === n)) {
              ((t += l.charAt(i >> 2)), (t += l.charAt((3 & i) << 4)), (t += "=="));
              break;
            }
            if (((r = e.charCodeAt(a++)), a === n)) {
              ((t += l.charAt(i >> 2)),
                (t += l.charAt(((3 & i) << 4) | ((240 & r) >> 4))),
                (t += l.charAt((15 & r) << 2)),
                (t += "="));
              break;
            }
            ((o = e.charCodeAt(a++)),
              (t += l.charAt(i >> 2)),
              (t += l.charAt(((3 & i) << 4) | ((240 & r) >> 4))),
              (t += l.charAt(((15 & r) << 2) | ((192 & o) >> 6))),
              (t += l.charAt(63 & o)));
          }
          return t;
        }
        var i = a ? "&type=" + a : "",
          r = document.location.protocol + "//nsclick.baidu.com/v.gif?pid=111&v=" + new Date().getTime() + i,
          o = "";
        for (var l in t) o = o + l + ":" + t[l] + ",";
        o = n("{" + o.substring(0, o.length - 1) + "}");
        for (var s in e) e.hasOwnProperty(s) && (r += "&" + s + "=" + e[s]);
        if (
          ((r += "&source=pc"),
          (r += "&auto_statistic=" + o + "&auto_en=" + t.eventType),
          t.errno && (r += "&errno=" + t.errno),
          r)
        ) {
          var c = new Image();
          ((c.onload = c.onerror =
            function () {
              ((c.onload = c.onerror = null), (c = null));
            }),
            (c.src = r));
        }
      },
      logClickLink: function (e, t) {
        ((e = e || {}), (e = this.formatLogData(e)));
        var a = t.target || t.srcElement,
          n = a && a.getAttribute && a.getAttribute("target"),
          i = "_self" === n || null === n,
          r = this.makeImgSrc(e);
        if ((i && (t.preventDefault ? t.preventDefault() : (t.returnValue = !1)), window.navigator.sendBeacon && i))
          (window.navigator.sendBeacon(r), (window.location.href = a.getAttribute("href")));
        else if (i) this.lazyLog(a, r, null);
        else {
          var o = new Image();
          ((o.onload = o.onerror =
            function () {
              ((o.onload = o.onerror = null), (o = null));
            }),
            (o.src = r));
        }
      },
      logCallBack: function (e, t) {
        ((e = e || {}), (e = this.formatLogData(e)));
        var a = this.makeImgSrc(e);
        this.lazyLog({}, a, t);
      },
      formatLogData: function (e) {
        var t = this.config || {};
        return baidu.extend(
          {},
          {
            page: e.logPage || "",
            source: "pc-api",
            tpl: t.product || "",
            subpro: t.subpro || "",
            extrajson: t.extrajson || "",
            data_source: "fe",
          },
          e,
        );
      },
      makeImgSrc: function (e) {
        var t = document.location.protocol + "//nsclick.baidu.com/v.gif?pid=111&type=1023&v=" + new Date().getTime(),
          a = "";
        for (var n in e) e.hasOwnProperty(n) && (a += "en" !== n ? "&" + n + "=" + e[n] : "&auto_en=" + e[n]);
        var i = "{eventType:" + (e.en || "") + "}";
        return ((i = this.base64encode(i)), (t += "&auto_statistic=" + i), (t += "&auto_en=" + (e.en || "")), (t += a));
      },
      lazyLog: function (e, t, a) {
        function n() {
          a ? a() : (window.location.href = i);
        }
        var i = e.getAttribute && e.getAttribute("href"),
          r = new Image(),
          o = setTimeout(function () {
            ((r.onload = r.onerror = r = null), n("imgTimeout"));
          }, 200);
        ((r.onload = r.onerror =
          function () {
            (clearTimeout(o), (r.onload = r.onerror = r = null), n("imgSuccess"));
          }),
          (r.src = t));
      },
      base64encode: function (e) {
        var t,
          a,
          n,
          i,
          r,
          o,
          l = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
        for (n = e.length, a = 0, t = ""; n > a;) {
          if (((i = 255 & e.charCodeAt(a++)), a === n)) {
            ((t += l.charAt(i >> 2)), (t += l.charAt((3 & i) << 4)), (t += "=="));
            break;
          }
          if (((r = e.charCodeAt(a++)), a === n)) {
            ((t += l.charAt(i >> 2)),
              (t += l.charAt(((3 & i) << 4) | ((240 & r) >> 4))),
              (t += l.charAt((15 & r) << 2)),
              (t += "="));
            break;
          }
          ((o = e.charCodeAt(a++)),
            (t += l.charAt(i >> 2)),
            (t += l.charAt(((3 & i) << 4) | ((240 & r) >> 4))),
            (t += l.charAt(((15 & r) << 2) | ((192 & o) >> 6))),
            (t += l.charAt(63 & o)));
        }
        return t;
      },
      _rendEventPlaceholder: function (e, t, a) {
        if (e || a) {
          var n,
            i,
            r = this,
            o = function (e) {
              1 === +e
                ? (t && t[0] && r.$hide(t[0]), a && a[0] && r.$show(a[0]))
                : (t && t[0] && r.$show(t[0]), a && a[0] && r.$hide(a[0]));
            };
          (setTimeout(function () {
            e && e.value && o(1);
          }, 200),
            baidu(t).on("mousedown", function () {
              ((i = !0),
                clearTimeout(n),
                (n = setTimeout(function () {
                  "password" === t[0] || (r.suggestionDom && "none" !== r.suggestionDom.style.display) || e.focus();
                }, 0)));
            }),
            baidu(a).on("click", function () {
              ((e.value = ""),
                o(0),
                e.focus(),
                r.suggestionDom &&
                  ((r.suggestionDom.data_delete = !0),
                  setTimeout(function () {
                    r.suggestionDom.data_delete = !1;
                  }, 200)));
            }),
            baidu(e).on("keyup", function () {
              o(e.value ? 1 : 0);
            }),
            baidu(e).on("focus", function () {
              window.inputCheckTimer = setInterval(function () {
                e.value.length ? (o(1), clearInterval(window.inputCheckTimer)) : o(0);
              }, 30);
            }));
        }
      },
      SBCtoDBC: function (e) {
        var t = "";
        if (e) {
          for (var a = e.length, n = 0; a > n; n++) {
            var i = e.charCodeAt(n);
            ((i = i >= 65281 && 65374 >= i ? i - 65248 : i), (i = 12288 == i ? 32 : i), (t += String.fromCharCode(i)));
          }
          return t;
        }
      },
      hide: function () {
        this.getElement().style.display = "none";
      },
      show: function () {
        this.getElement().style.display = "block";
      },
      _analysis: function (e, t) {
        return window.passport.analysis &&
          window.passport.analysis[this.module] &&
          window.passport.analysis[this.module][e]
          ? (window.passport.analysis[this.module][e](this, t), { preventDefault: !1, preventEvent: !1 })
          : void 0;
      },
      _hook: function (e, t) {
        return window.passport.hook && window.passport.hook[this.module] && window.passport.hook[this.module][e]
          ? window.passport.hook[this.module][e](this, t)
          : { preventDefault: !1, preventEvent: !1 };
      },
      fireEvent: function (e, t) {
        var a = this._hook(e, t),
          n = !0;
        return (a.preventEvent || (n = this.fire(e, t)), !a.preventDefault && n);
      },
      getQueryString: function (e) {
        var t = new RegExp("(^|&)" + e + "=([^&]*)(&|$)", "i"),
          a = window.location.search.substr(1).match(t);
        return null != a ? unescape(a[2]) : "";
      },
    });
  var hexcase = 0;
  var b64pad = "";
  var chrsz = 8;
  function hex_md5(s) {
    return binl2hex(core_md5(str2binl(s), s.length * chrsz));
  }
  function b64_md5(s) {
    return binl2b64(core_md5(str2binl(s), s.length * chrsz));
  }
  function str_md5(s) {
    return binl2str(core_md5(str2binl(s), s.length * chrsz));
  }
  function hex_hmac_md5(key, data) {
    return binl2hex(core_hmac_md5(key, data));
  }
  function b64_hmac_md5(key, data) {
    return binl2b64(core_hmac_md5(key, data));
  }
  function str_hmac_md5(key, data) {
    return binl2str(core_hmac_md5(key, data));
  }
  function md5_vm_test() {
    return hex_md5("abc") == "900150983cd24fb0d6963f7d28e17f72";
  }
  function core_md5(x, len) {
    x[len >> 5] |= 128 << (len % 32);
    x[(((len + 64) >>> 9) << 4) + 14] = len;
    var a = 1732584193;
    var b = -271733879;
    var c = -1732584194;
    var d = 271733878;
    for (var i = 0; i < x.length; i += 16) {
      var olda = a;
      var oldb = b;
      var oldc = c;
      var oldd = d;
      a = md5_ff(a, b, c, d, x[i + 0], 7, -680876936);
      d = md5_ff(d, a, b, c, x[i + 1], 12, -389564586);
      c = md5_ff(c, d, a, b, x[i + 2], 17, 606105819);
      b = md5_ff(b, c, d, a, x[i + 3], 22, -1044525330);
      a = md5_ff(a, b, c, d, x[i + 4], 7, -176418897);
      d = md5_ff(d, a, b, c, x[i + 5], 12, 1200080426);
      c = md5_ff(c, d, a, b, x[i + 6], 17, -1473231341);
      b = md5_ff(b, c, d, a, x[i + 7], 22, -45705983);
      a = md5_ff(a, b, c, d, x[i + 8], 7, 1770035416);
      d = md5_ff(d, a, b, c, x[i + 9], 12, -1958414417);
      c = md5_ff(c, d, a, b, x[i + 10], 17, -42063);
      b = md5_ff(b, c, d, a, x[i + 11], 22, -1990404162);
      a = md5_ff(a, b, c, d, x[i + 12], 7, 1804603682);
      d = md5_ff(d, a, b, c, x[i + 13], 12, -40341101);
      c = md5_ff(c, d, a, b, x[i + 14], 17, -1502002290);
      b = md5_ff(b, c, d, a, x[i + 15], 22, 1236535329);
      a = md5_gg(a, b, c, d, x[i + 1], 5, -165796510);
      d = md5_gg(d, a, b, c, x[i + 6], 9, -1069501632);
      c = md5_gg(c, d, a, b, x[i + 11], 14, 643717713);
      b = md5_gg(b, c, d, a, x[i + 0], 20, -373897302);
      a = md5_gg(a, b, c, d, x[i + 5], 5, -701558691);
      d = md5_gg(d, a, b, c, x[i + 10], 9, 38016083);
      c = md5_gg(c, d, a, b, x[i + 15], 14, -660478335);
      b = md5_gg(b, c, d, a, x[i + 4], 20, -405537848);
      a = md5_gg(a, b, c, d, x[i + 9], 5, 568446438);
      d = md5_gg(d, a, b, c, x[i + 14], 9, -1019803690);
      c = md5_gg(c, d, a, b, x[i + 3], 14, -187363961);
      b = md5_gg(b, c, d, a, x[i + 8], 20, 1163531501);
      a = md5_gg(a, b, c, d, x[i + 13], 5, -1444681467);
      d = md5_gg(d, a, b, c, x[i + 2], 9, -51403784);
      c = md5_gg(c, d, a, b, x[i + 7], 14, 1735328473);
      b = md5_gg(b, c, d, a, x[i + 12], 20, -1926607734);
      a = md5_hh(a, b, c, d, x[i + 5], 4, -378558);
      d = md5_hh(d, a, b, c, x[i + 8], 11, -2022574463);
      c = md5_hh(c, d, a, b, x[i + 11], 16, 1839030562);
      b = md5_hh(b, c, d, a, x[i + 14], 23, -35309556);
      a = md5_hh(a, b, c, d, x[i + 1], 4, -1530992060);
      d = md5_hh(d, a, b, c, x[i + 4], 11, 1272893353);
      c = md5_hh(c, d, a, b, x[i + 7], 16, -155497632);
      b = md5_hh(b, c, d, a, x[i + 10], 23, -1094730640);
      a = md5_hh(a, b, c, d, x[i + 13], 4, 681279174);
      d = md5_hh(d, a, b, c, x[i + 0], 11, -358537222);
      c = md5_hh(c, d, a, b, x[i + 3], 16, -722521979);
      b = md5_hh(b, c, d, a, x[i + 6], 23, 76029189);
      a = md5_hh(a, b, c, d, x[i + 9], 4, -640364487);
      d = md5_hh(d, a, b, c, x[i + 12], 11, -421815835);
      c = md5_hh(c, d, a, b, x[i + 15], 16, 530742520);
      b = md5_hh(b, c, d, a, x[i + 2], 23, -995338651);
      a = md5_ii(a, b, c, d, x[i + 0], 6, -198630844);
      d = md5_ii(d, a, b, c, x[i + 7], 10, 1126891415);
      c = md5_ii(c, d, a, b, x[i + 14], 15, -1416354905);
      b = md5_ii(b, c, d, a, x[i + 5], 21, -57434055);
      a = md5_ii(a, b, c, d, x[i + 12], 6, 1700485571);
      d = md5_ii(d, a, b, c, x[i + 3], 10, -1894986606);
      c = md5_ii(c, d, a, b, x[i + 10], 15, -1051523);
      b = md5_ii(b, c, d, a, x[i + 1], 21, -2054922799);
      a = md5_ii(a, b, c, d, x[i + 8], 6, 1873313359);
      d = md5_ii(d, a, b, c, x[i + 15], 10, -30611744);
      c = md5_ii(c, d, a, b, x[i + 6], 15, -1560198380);
      b = md5_ii(b, c, d, a, x[i + 13], 21, 1309151649);
      a = md5_ii(a, b, c, d, x[i + 4], 6, -145523070);
      d = md5_ii(d, a, b, c, x[i + 11], 10, -1120210379);
      c = md5_ii(c, d, a, b, x[i + 2], 15, 718787259);
      b = md5_ii(b, c, d, a, x[i + 9], 21, -343485551);
      a = safe_add(a, olda);
      b = safe_add(b, oldb);
      c = safe_add(c, oldc);
      d = safe_add(d, oldd);
    }
    return Array(a, b, c, d);
  }
  function md5_cmn(q, a, b, x, s, t) {
    return safe_add(bit_rol(safe_add(safe_add(a, q), safe_add(x, t)), s), b);
  }
  function md5_ff(a, b, c, d, x, s, t) {
    return md5_cmn((b & c) | (~b & d), a, b, x, s, t);
  }
  function md5_gg(a, b, c, d, x, s, t) {
    return md5_cmn((b & d) | (c & ~d), a, b, x, s, t);
  }
  function md5_hh(a, b, c, d, x, s, t) {
    return md5_cmn(b ^ c ^ d, a, b, x, s, t);
  }
  function md5_ii(a, b, c, d, x, s, t) {
    return md5_cmn(c ^ (b | ~d), a, b, x, s, t);
  }
  function core_hmac_md5(key, data) {
    var bkey = str2binl(key);
    if (bkey.length > 16) {
      bkey = core_md5(bkey, key.length * chrsz);
    }
    var ipad = Array(16),
      opad = Array(16);
    for (var i = 0; i < 16; i++) {
      ipad[i] = bkey[i] ^ 909522486;
      opad[i] = bkey[i] ^ 1549556828;
    }
    var hash = core_md5(ipad.concat(str2binl(data)), 512 + data.length * chrsz);
    return core_md5(opad.concat(hash), 512 + 128);
  }
  function safe_add(x, y) {
    var lsw = (x & 65535) + (y & 65535);
    var msw = (x >> 16) + (y >> 16) + (lsw >> 16);
    return (msw << 16) | (lsw & 65535);
  }
  function bit_rol(num, cnt) {
    return (num << cnt) | (num >>> (32 - cnt));
  }
  function str2binl(str) {
    var bin = Array();
    var mask = (1 << chrsz) - 1;
    for (var i = 0; i < str.length * chrsz; i += chrsz) {
      bin[i >> 5] |= (str.charCodeAt(i / chrsz) & mask) << (i % 32);
    }
    return bin;
  }
  function binl2str(bin) {
    var str = "";
    var mask = (1 << chrsz) - 1;
    for (var i = 0; i < bin.length * 32; i += chrsz) {
      str += String.fromCharCode((bin[i >> 5] >>> (i % 32)) & mask);
    }
    return str;
  }
  function binl2hex(binarray) {
    var hex_tab = hexcase ? "0123456789ABCDEF" : "0123456789abcdef";
    var str = "";
    for (var i = 0; i < binarray.length * 4; i++) {
      str +=
        hex_tab.charAt((binarray[i >> 2] >> ((i % 4) * 8 + 4)) & 15) +
        hex_tab.charAt((binarray[i >> 2] >> ((i % 4) * 8)) & 15);
    }
    return str;
  }
  function binl2b64(binarray) {
    var tab = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    var str = "";
    for (var i = 0; i < binarray.length * 4; i += 3) {
      var triplet =
        (((binarray[i >> 2] >> (8 * (i % 4))) & 255) << 16) |
        (((binarray[(i + 1) >> 2] >> (8 * ((i + 1) % 4))) & 255) << 8) |
        ((binarray[(i + 2) >> 2] >> (8 * ((i + 2) % 4))) & 255);
      for (var j = 0; j < 4; j++) {
        if (i * 8 + j * 6 > binarray.length * 32) {
          str += b64pad;
        } else {
          str += tab.charAt((triplet >> (6 * (3 - j))) & 63);
        }
      }
    }
    return str;
  }
  /* globals passport, magic, confirmSmsVerifyWidget */
  /* eslint-disable max-len */
  /* eslint-disable */
  magic.passport.login = baidu.lang
    .createClass(
      function (options) {
        var me = this;
        if (passport && passport._protocol === "https") {
          var protocol = "https:";
        } else {
          var protocol = window.location
            ? window.location.protocol.toLowerCase()
            : document.location.protocol.toLowerCase();
        }

        me._domain = {
          https: "https://passport.baidu.com",
          http: "http://passport.baidu.com",
          auto: protocol === "https:" ? "https://passport.baidu.com" : "http://passport.baidu.com",
          wapAuto: protocol === "https:" ? "https://wappass.baidu.com" : "http://wappass.baidu.com",
        };
        me.config = {
          memberPass: true, // bool,默认是否记住用户名
          isQuickUser: 0, // 0,1 是否为快推登录
          safeFlag: 0,
          product: "", // string,产品线TPL
          hideEmailAndPhoneText: false, //隐藏邮箱和手机号文案
          idc: "", // string,session机房标记
          charset: "",
          loginMerge: true, // bool,是否为合并普通登录与手机号登录入口
          staticPage: "", // string,静态跳转文件
          hasRegUrl: false, // bool,是否有注册入口
          u: "",
          lang: "",
          autosuggest: false, // bool,是否有自动历史纪录及邮箱提示
          hasPlaceholder: false, // bool,是否有placeholder功能
          registerLink: "", // string,注册连接配置
          authsiteLogin: "", // array,第三方账号配置,目前支持['qzone', 'tsina', 'weixin']
          authsiteCfgLogin: "", // object,第三方详细配置,
          authsiteCurrentOpen: false, // 第三方登录打开方式: false为新开页面，true为当前页面
          appKey: "",
          qrcode: true, // bool,是否使用二维码登录
          sms: 0, // sms配置为5表示支持短信登录
          uniqueId: false, // bool,过渡参数,是否使用唯一的ID
          autoFocus: true, // bool,是否自动focus到焦点
          subpro: "", // 待定参数，用以统计
          setWebToClient: false, // 是否同步登录状态到客户端      ,
          is_voice_sms: 0,
          voice_sms_flag: 0,
          userPwdLogin: 0, // 默认账号密码登录
          qrcode_animation: 1,
          /* eslint-disable fecs-camelcase */
          qrcode_style: true,
          // 图形验证码触发的来源 username or login or '', 用户名失焦、单击登录、没有触发验证码
          vcodefrom: "",
          client: "",
          // 【爱番番】免登录跳转至百家号登录支持
          smsAutoLogin: {
            mobile: "",
            encryption: "",
          },
          // oauth h5扫码登录来源
          qrloginfrom: "pc",
          // 是否需要强制验证
          mustVerify: "",
          // 因yy多加一个clientId用于accesstoken接口请求
          clientId: "",
          qrcodeCfg: {},
          // 大搜打点需求，增加
          sid: "",
          // 允许业务线在加点上增加打点数据
          logParams: {},
          // 是否同步内网域名(baidu-int.com)bduss
          isSyncBaiduInt: "",
          // 是否展示儿童监护人协议
          showChildAgree: 0,
          // 是否限制互联网手机号
          inTMobile: 0,
          // 是否展示协议勾选框 1 展示 0 不展示，前提条件 showAgreement为 1
          showAgreeCheckbox: 0,
          maskId: "",
          fileId: "",
          certOptions: {},
          // 是否支持三方登录成功后回调业务方 通过postMessage
          canMessageChannel: 0,
          // 是否开启此客户端互通功能
          isOpenClientShare: false,
          // 配置客户端互通选项
          clientShareOpt: "",
          // 客户端互通是否有登录态
          clientShareIsLogin: false,
          // 客户端互通登录用户信息
          clientShareUserInfo: {},
          // 是否开启微信小程序二维码登录，仅composeTemplate为1/true时生效
          isOpenWxMiniQrcodeLogin: false,
          // 当前正处于哪个tab
          currentTab: "",
          // 自定义协议
          customAgreementList: [],
          // 是否开启三方登录协议校验弹窗 1 开启 0 不开启
          thirdCheckAgreement: 0,
          // 扫码登录二维码携带的隐私链接
          qrcodePrivacyLink: "",
          // 开启切换海外登录入口
          isOpenOverseasLogin: false,
          // 海外登录配置
          overseasLoginOpt: {},
        };
        baidu.extend(me.config, options);
        me._setPloganondegCookie();
        me.urlData = {
          page: "loginv5",
          tpl: me.config.product || "",
          subpro: me.config.subpro || "",
          traceId: "pc_loginv5_" + (new Date().getTime() / 1000).toFixed(),
        };
        if (me.config.sid) {
          me.urlData.sid = me.config.sid;
        }
        if (me.config.logParams && me.config.logParams instanceof Object) {
          for (var keys in me.config.logParams) {
            me.urlData[keys] = me.config.logParams[keys];
          }
        }
        me.config.product = me.config.product || "isnull";
        me.config.qrcode = me.config.qrcode === 0 ? 0 : 3;
        me.config.page = "loginv5";
        me.config.hideEmailAndPhoneText = me.config.hideEmailAndPhoneText || false;
        me.config.loginType = "pwdLogin";
        // 必展示协议勾选框的业务线(首次监管查)
        var showAgreeCheckboxTpl = ["mn", "search_aichat", "ma", "do", "netdisk"];
        // 最终是否展示协议
        me.config.showAgreeCheckbox =
          me.config.showAgreeCheckbox || showAgreeCheckboxTpl.indexOf(me.config.product) !== -1;

        this.module = "login";
        this.guideRandom = (function () {
          return "xxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
            .replace(/[xy]/g, function (c) {
              var r = (Math.random() * 16) | 0;
              var v = c === "x" ? r : (r & 0x3) | 0x8;
              return v.toString(16);
            })
            .toUpperCase();
        })();
        // 从登录跳转到单独reg页面，繁体支持
        var currentQueryLang = me.getQueryString("lang") || me.config.lang;
        var langQueryString = currentQueryLang ? "&lang=" + currentQueryLang : "";
        var langQueryFirstString = currentQueryLang ? "?lang=" + currentQueryLang : "";
        me.tplQuery = me.config.product || passport.tpl || me.getQueryString("tpl") || "";
        me.currentQueryLang = currentQueryLang;
        me.isEnVersion = passport.enTpls && passport.enTpls.includes(me.tplQuery);
        me.constant = {
          CHECKVERIFYCODE: true,
          CONTAINER_CLASS: "tang-pass-login",
          LABEL_FOCUS_CLASS: "pass-text-label-focus",
          FOCUS_CLASS: "pass-text-input-focus",
          HOVER_CLASS: "pass-text-input-hover",
          ERROR_CLASS: "pass-text-input-error",
          VERIFYCODE_URL_PREFIX: me._domain.https + "/cgi-bin/genimage?",
          BLANK_IMG_URL: passport.apiDomain.staticDomain + "/passApi/img/small_blank.gif",
          MODIFY_PWD_URL_PREFIX: me._domain.https + "/forcsdnpasschange",
          GET_PASSWORD_URL:
            me.config.forgetLink ||
            me._domain.https +
              "/?getpassindex&tt=" +
              new Date().getTime() +
              "&gid=" +
              me.guideRandom +
              "&tpl=" +
              encodeURIComponent(me.config.product) +
              "&u=" +
              encodeURIComponent(me.config.u) +
              langQueryString,
          REG_URL:
            me.config.registerLink ||
            me._domain.https +
              "/v2/?reg&tt=" +
              new Date().getTime() +
              "&overseas=" +
              me.config.overseas +
              "&gid=" +
              me.guideRandom +
              "&tpl=" +
              encodeURIComponent(me.config.product) +
              "&u=" +
              encodeURIComponent(me.config.u) +
              langQueryString,
          PROTOCAL_URL: me._domain.http + "/static/passpc-account/html/protocal.html" + langQueryFirstString,
          PERSONAL_PROTOCAL_URL: "https://privacy.baidu.com/policy/PrivacyPolicy",
          CHILD_AGREEMENT_URL: "https://privacy.baidu.com/policy/children-privacy-policy/index.html?_" + Date.now(),
          NOCAPTCHA_URL: me._domain.auto + "/static/passpc-base/js/ld.min.js?cdnversion=" + new Date().getTime(),
          TOUCHAPIMKD_URL: me._domain.wapAuto + "/static/machine/js/api/mkd.js?cdnversion=" + new Date().getTime(),
          SUBMITFLAG: false, // 给游戏提供的登录接口判断
          TOUCHAPIMKD_URL_V2:
            me._domain.wapAuto + "/static/sdk-machine/js/mkd_v2.js?cdnversion=" + new Date().getTime(),
          RESETPASSWORD_URL: me._domain.auto + "/passApi/js//passResetpwdSdk.js?cdnversion=" + new Date().getTime(),
          OVERSEASLOGIN_URL:
            me._domain.auto + "/passApi/js//passOverseasLoginSdk.js?cdnversion=" + new Date().getTime(),
          UPSMSSDK_URL: me._domain.auto + "/passApi/js//passUpSmsSdk.js?cdnversion=" + new Date().getTime(),
        };
        me.config.protocalLink && (me.constant.PROTOCAL_URL = me.config.protocalLink);
        me.config.userpersonalLink && (me.constant.PERSONAL_PROTOCAL_URL = me.config.userpersonalLink);
        me.config.regLink && (me.constant.REG_URL = me.config.regLink);
        me.lang = passport.err.getCurrent(me.config.lang).labelText.login;
        me.PcDukeUrl = passport && passport.PcDukeUrl;
        me.hooksText = passport.err.getCurrent(me.config.lang).hooksText;
        me.config.customAgreementList = me.config.customAgreementList && me.config.customAgreementList.slice(0, 3);
        me.customizeAnd = me.config.customAgreementList.length !== 0;

        // init data
        passport.data.setContext(baidu.extend({}, me.config));
        this.initialized = false;
        this.validatorInited = false;
        this.bdPsWtoken = "";
        this.innerData = { normal: {}, phone: {} };
        this.dataFiels = ["userName", "password"];
        this.initTime = new Date().getTime();
        window.confirmSmsVerifyWidget = null;
        window.checkPhoneWidget = null;
        window.checkPhoneExist = false;

        var passLowerIE = /msie (\d+\.\d+)/i.test(navigator.userAgent)
          ? document.documentMode || +RegExp["\x241"]
          : undefined;
        if (passLowerIE <= 6) {
          me.config.qrcode_animation = false;
        }
        me.passLowerIE = passLowerIE;

        if (me.config.overseas && me.config.overseas == 1) {
          this.foreignMobile = true;
        }
        this.internation = false;
        me._insertNoCaptchaScript();

        // 新人机
        me.initMkd();
        me.initMkdFlag = false;
        me._insertScriptW(me.constant.TOUCHAPIMKD_URL_V2, function () {
          me.initMkdv2();
        });

        // 是否加载默认CSS
        if (me.config.defaultCss) {
          me._loadCssFileW("/passApi/css/loginv5_c9af7d5.css", function () {});
        }
        // 用于保证接入业务线不加载默认样式的情况----拓展节点保证样式
        me._loadCssFileW("/passApi/css/login_common_7a336df.css", function () {});

        me._insertScriptW(
          "https://wappass.baidu.com/static/touch/js/lib/fingerprint.js?tt=" + +new Date(),
          function () {
            me.fuid = window.passFingerPrint();
            me._setCookie("ppfuid", me.fuid && me.fuid.fuid, {
              domain: ".baidu.com",
              secure: true,
            });
          },
        );

        // 海外登录sdk
        if (me.config.isOpenOverseasLogin) {
          me.config.overseasLoginOpt = Object.assign(
            {
              // 轮播图，最多支持3张
              bannerImages: [],
              tpl: "",
              // 自定义协议
              agreements: [],
              // 三方登录方式 google
              thirdLoginTypes: [],
              // 关闭弹窗回调
              onClose: function () {},
              // 切换国内回调
              onSwitchToChinaVersion: function () {},
              // 渲染完成回调
              renderCallback: function () {},
              // 主题
              themeColor: "#4E6EF2",
              // 登录成功回跳地址
              u: "",
              // 自定义文案
              customText: {
                switchToOverseas: "",
                switchToChina: "",
              },
            },
            options.overseasLoginOpt || {},
          );
          me._insertScriptW(me.constant.OVERSEASLOGIN_URL);
        }

        // getapi:针对不愿意配置memberPass（自动登录）并且需要调用getapi接口时候配置
        if (me.config.memberPass || me.config.getapi) {
          me._initApi();
        }
        if (me.config.sms + "" === "5") {
          this.smsConfig = true;
        } else {
          this.smsConfig = false;
        }

        // 如果开启微信小程序二维码登录
        if (me.config.isOpenWxMiniQrcodeLogin) {
          // 如果为开启复合面板，则关闭微信小程序二维码登录
          !me.config.composeTemplate && (me.config.isOpenWxMiniQrcodeLogin = false);
          me.config.qrcode = 0;
          me.config.qrcodeLogin = 0;
        }
      },
      {
        type: "magic.passport.login",
        superClass: magic.passport,
      },
    )
    .extend({
      // 点击行为回调
      _callBehaviorCallback: function (data) {
        try {
          var me = this;
          me.config &&
            me.config.onBehaviorCallback &&
            typeof me.config.onBehaviorCallback === "function" &&
            me.config.onBehaviorCallback(data);
        } catch (error) {
          console.error("Behavior callback error:", error);
        }
      },
      _setComposeTemplate: function () {
        var composeTemplateStr =
          '<div id="componseLeft"  class="compose-left-container"></div>' +
          '<div id="componseRight" class="compose-right-container"></div>';
        return composeTemplateStr;
      },
      _getIrregularField: function (field) {
        var me = this;
        var template = {
          makeText: '<p id="' + me.$getId("MakeTextWrapper") + '" class="pass-make-text" style="display:none;"></p>',
          verifyCode:
            '<p id="' +
            me.$getId("verifyCodeImgWrapper") +
            '" class="pass-form-item pass-form-item-verifyCode" style="display:none">' +
            '<input id="' +
            me.$getId("verifyCode") +
            '" type="text" name="verifyCode" class="pass-text-input pass-text-input-verifyCode" maxlength="6" />' +
            '<span  id="' +
            me.$getId("verifyCodeImgParent") +
            '" class="pass-verifyCodeImgParent" ><img id="' +
            me.$getId("verifyCodeImg") +
            '" class="pass-verifyCode" src="' +
            me.constant.BLANK_IMG_URL +
            '" /></span>' +
            '<a id="' +
            me.$getId("verifyCodeChange") +
            '" href="#" class="pass-change-verifyCode">' +
            me.lang.captchaChange +
            "</a>" +
            '<span id="' +
            me.$getId("verifyCodeError") +
            '" class="pass-error pass-error-verifyCode"></span>' +
            '<span id="' +
            me.$getId("verifyCodeTip") +
            '" class="pass-tip pass-tip-verifyCode"></span>' +
            '<span id="' +
            me.$getId("verifyCodeSuccess") +
            '" class="pass-success pass-success-verifyCode"></span>' +
            "</p>",
          unbindingSMSTips:
            '<p id="' +
            me.$getId("unbindingSMSTipsWrapper") +
            '" class="pass-unbindingSMSTipsWrapper" style="display: none;">' +
            me.lang.unbindingSmsTip +
            '<span id="' +
            me.$getId("unbindingSMSTipsBtn") +
            '" class="pass-unbindingSMSTipsBtn">' +
            me.lang.unblocking +
            "</span></p>",
          certErrorTips:
            '<p id="' +
            me.$getId("certErrorWrapper") +
            '" class="pass-unbindingSMSTipsWrapper" style="display: none; margin-bottom: 14px;">' +
            me.lang.unbindingSmsTip1 +
            '<span id="' +
            me.$getId("certErrorBtn") +
            '" class="pass-unbindingSMSTipsBtn">' +
            me.lang.unblocking +
            "</span></p>",
          generalError:
            '<p id="' +
            me.$getId("errorWrapper") +
            '" class="pass-generalErrorWrapper">' +
            '<span id="' +
            me.$getId("error") +
            '" class="pass-generalError pass-generalError-error"></span>' +
            "</p>",
          rem:
            '<p id="' +
            me.$getId("memberPassWrapper") +
            '" class="pass-form-item pass-form-item-memberPass memberPass-hidden">' +
            '<input id="' +
            me.$getId("memberPass") +
            '" type="checkbox" name="memberPass" class="pass-checkbox-input pass-checkbox-memberPass"' +
            (me.config.memberPass ? ' checked="checked"' : "") +
            " />" +
            '<label for="' +
            me.$getId("memberPass") +
            '" id="' +
            me.$getId("memberPassLabel") +
            '" class="">' +
            me.lang.memberPassLabel +
            "</label>" +
            "</p>",
          submit:
            '<p id="' +
            me.$getId("submitWrapper") +
            '" class="pass-form-item pass-form-item-submit">' +
            '<input id="' +
            me.$getId("submit") +
            '" type="submit" value="' +
            me.lang.login +
            '" class="pass-button pass-button-submit' +
            (me.config.showAgreeCheckbox ? ' pass-button-full-disabled" disabled />' : '"/>') +
            '<span id="' +
            me.$getId("agreementWrapper") +
            '" class="tang-pass-sms-agreement pass-link">' +
            (me.config.showAgreeCheckbox
              ? '<input name="isAgree" id="' +
                me.$getId("isAgree") +
                '" type="checkbox" class="pass-checkbox-input pass-checkbox-isAgree" autocomplete="off" />' +
                '<label for="' +
                me.$getId("isAgree") +
                '" style="padding-left: 5px;">' +
                me.lang.agree +
                "</label>"
              : me.lang.agree) +
            '<a target="_blank" href="' +
            me.constant.PROTOCAL_URL +
            '" style="display: inline-block;margin-bottom: 6px">' +
            (me.config.protocalText || me.lang.baiduUserProtocal) +
            "</a> " +
            (me.customizeAnd ? "、" : me.lang.and) +
            '<a target="_blank" href="' +
            me.constant.PERSONAL_PROTOCAL_URL +
            '">' +
            (me.config.userpersonalText || me.lang.baiduUserPersonal) +
            "</a>" +
            (me.config.showChildAgree
              ? '、<a target="_blank" href="' +
                me.constant.CHILD_AGREEMENT_URL +
                '">' +
                me.lang.childAgreement +
                "</a> "
              : "") +
            "</span>" +
            // + (me.smsConfig ? '<a class="pass-sms-btn pass-link" title="' + me.lang.toSmsBtn
            //     + '" data-type="sms" id="' + me.$getId('smsSwitchWrapper') + '">' + me.lang.toSmsBtn + '</a>' : '')
            "</p>",
          forgetPwd: '<a class="pass-fgtpwd pass-link" id="' + me.$getId("fgtpwdLink") + '">' + me.lang.fgtPwd + "</a>",
          foreignMobileMsg:
            '<p class="pass-foreignMobile-msg" id="' +
            me.$getId("foreignMobileMsg") +
            '">' +
            me.lang.foreignMobileMsg +
            "</p>",
          foreignMobileWrapper:
            '<div class="pass-form-item pass-form-item-PhoneCountry pass-foreignMobile-wrapper" id="' +
            me.$getId("foreignMobileWrapper") +
            '" style="display:none">' +
            '<label for="' +
            me.$getId("foreignMobile") +
            '" id="' +
            me.$getId("foreignMobileLabel") +
            '" class="pass-label" data-countryCode="' +
            (me.isEnVersion ? "001" : "") +
            '">' +
            (me.isEnVersion ? "+1" : "+86") +
            "</label>" +
            '<input id="' +
            me.$getId("foreignMobile") +
            '" type="text" name="foreignusername" class="pass-text-input"/>' +
            '<ul id="' +
            me.$getId("foreignCountryList") +
            '" class="pass-country-list"></ul>' +
            "</div>",
          foreignMobileLink:
            '<p class="pass-foreignMobile-link-wrapper" id="' +
            me.$getId("foreignMobileLinkWrapper") +
            '"><a id="' +
            me.$getId("foreignMobileLink") +
            '" class="pass-foreignMobile-link">' +
            me.lang.foreignMobileLink +
            "</a></p>",
          foreignMobileBack:
            '<p class="pass-foreignMobile-back-wrapper" id="' +
            me.$getId("foreignMobileBackWrapper") +
            '"><a id="' +
            me.$getId("foreignMobileBack") +
            '" class="pass-foreignMobile-link">' +
            me.lang.foreignToLogin +
            "</a></p>",
          choiceuser:
            '<div id="' +
            me.$getId("choiceuser_article") +
            '" class="tang-pass-login-choice choiceuser-article">' +
            '<div class="choiceuser-msg"><p class="choiceuser-msg-title">' +
            me.lang.dearUser +
            "</p>" +
            '<p class="choiceuser-msg-text"> ' +
            me.lang.choiceuserTip +
            " ：</p></div>" +
            '<div class="choiceuser-buttons"><div class="choiceuser-btn">' +
            '<input class="pass-button pass-button-choiceuser-username" type="button" value="' +
            me.lang.userNameTitle +
            '" id="' +
            me.$getId("choiceuser_btn_username") +
            '"/>' +
            '<input class="pass-button pass-button-choiceuser-phone" type="button" value="' +
            me.lang.phoneNum +
            '" id="' +
            me.$getId("choiceuser_btn_mobile") +
            '"/></div>' +
            '<div class="choiceuser-back"><a href="#" id="' +
            me.$getId("choiceuser_btn_back") +
            '">' +
            me.lang.backToLogin +
            "</a></div></div>" +
            "</div>",
          webtoclint: '<div id="' + me.$getId("pass_b2c") + '" style="display:none;"></div>',
          is_voice_sms:
            '<div class="pass-smsSwitchWrapper"><a class="pass-is_voice_sms-btn" title="' +
            me.lang.vCode +
            '" data-type="is_voice_sms" >' +
            me.lang.vCode +
            "</a></div>",
          upsms_dialog:
            '<div class="upsms-dialog-wrap" id="' +
            me.$getId("upsms-dialog-wrap") +
            '" style="display:none;">' +
            '<div class="upsms-dialog-box">' +
            '<div class="upsms-dialog-header">' +
            '<h5 class="upsms-dialog-header-title">' +
            me.lang.securityTips +
            "</h5>" +
            "</div>" +
            '<div class="upsms-dialog-body">' +
            '<p  class="upsms-dialog-body-txt">' +
            me.lang.upsmsTip +
            '<span id="' +
            me.$getId("upsms-dialog-phone") +
            '"></span>发送短信<span id="' +
            me.$getId("upsms-dialog-code") +
            '"></span>到<span id="' +
            me.$getId("upsms-dialog-smsNumber") +
            '"></span>' +
            me.lang.continueToValidate +
            "</p>" +
            "</div>" +
            '<div class="upsms-dialog-footer">' +
            '<button class="upsms-dialog-footer-btn" id="' +
            me.$getId("upsms-dialog-btn") +
            '">' +
            me.lang.confirmText +
            "</button>" +
            "</div>" +
            "</div>" +
            "</div>",
          tipToast:
            '<div class="tipToast-wrap" id="' +
            me.$getId("tipToast-wrap") +
            '"><p class="tipToast-body" id="' +
            me.$getId("tipToast-body") +
            '"></p></div>',
        };
        return template[field];
      },
      _getHeaderTemplate: function () {
        var me = this;
        var headerTemplateStr =
          (me.config.composeTemplate && me.config.hideLogo
            ? '<p class="hide-logo-text">' + me.lang.qrcodeText + "</p>"
            : me.config.hideLogo
              ? '<p class="hide-logo-not-text"></p>'
              : '<p class="pass-form-logo"></p>') +
          '<p id="' +
          me.$getId("headerLoginTab") +
          '" class="login-type-tab">' +
          '<span id="' +
          me.$getId("changeQrCodeItem") +
          '" class="switch-item"' +
          (me.config.composeTemplate || (me.config.qrcode === 0 && me.config.qrcodeLogin === 0)
            ? 'style="display: none;"'
            : "") +
          '">' +
          me.lang.qrcodeBtn +
          "</span>" +
          '<span id="' +
          me.$getId("changePwdCodeItem") +
          '" class="switch-item">' +
          me.lang.clientOnekeyChangeLogin +
          "</span>" +
          '<span id="' +
          me.$getId("changeSmsCodeItem") +
          '" class="switch-item"' +
          (!me.smsConfig ? 'style="display: none;"' : "") +
          ">" +
          me.lang.smsLogin +
          "</span>" +
          "</p>";
        return headerTemplateStr;
      },
      _getTemplate: function (containerId) {
        var me = this;
        var templateStr =
          '<form id="' + me.$getId("form") + '" class="pass-form pass-form-normal" method="POST" autocomplete="off">';
        var hiddenFields = {
          codeString: "",
          safeFlag: me.config.safeFlag,
          u: me.config.u,
          isPhone: false,
          detect: "1",
          gid: me.guideRandom || "",
          staticPage: me.config.staticPage,
          quick_user: me.config.isQuickUser, // 注意产品线配置时以驼峰命名
          logintype: me.config.diaPassLogin ? "dialogLogin" : "basicLogin", // 用以log统计
          logLoginType: me.config.diaPassLogin ? "pc_loginDialog" : "pc_loginBasic", // 用以zmon监控统计
          subpro: me.config.subpro,
          idc: me.config.idc,
          loginMerge: true,
          mkey: me.config.mkey || "",
        };
        var loginRegularField = [
          {
            field: "userName",
            noError: true,
          },
          {
            field: "password",
            pwd: true,
            noError: true,
          },
        ];
        // templateStr += me._getIrregularField('logoText');
        // templateStr += me._getIrregularField('loginTypeTab');
        templateStr += me._getIrregularField("makeText");
        templateStr += me._getHiddenField(hiddenFields);
        if (me.foreignMobile) {
          templateStr += me._getIrregularField("foreignMobileMsg");
          templateStr += me._getIrregularField("foreignMobileWrapper");
        }

        for (var i = 0; i < loginRegularField.length; i++) {
          templateStr += me._getRegularField(loginRegularField[i]);
        }
        templateStr += me._getIrregularField("verifyCode");
        templateStr += me._getIrregularField("rem");
        if (me.foreignMobile && !me.isEnVersion) {
          templateStr += me._getIrregularField("foreignMobileLink");
          templateStr += me._getIrregularField("foreignMobileBack");
        }
        templateStr += me._getIrregularField("generalError");
        templateStr += me._getIrregularField("unbindingSMSTips");
        templateStr += me._getIrregularField("certErrorTips");

        templateStr += me._getIrregularField("forgetPwd");
        templateStr += me._getIrregularField("submit");
        templateStr += "</form>";
        return templateStr;
      },
      /**
       * 渲染客户端互通一键登录模版
       * @returns
       */
      _getClientShareTemplate: function () {
        var me = this;
        var config = me.config;
        var templateStr =
          '<div id="' +
          me.$getId("clientonekey") +
          '" class="pass-clientonekey-box">' +
          '<p class="clientonekey-tip">' +
          me.lang.clientOnekeytip1 +
          config.clientShareOpt.clientName +
          me.lang.clientOnekeytip2 +
          "</p>" +
          '<img id="' +
          me.$getId("clientOnekeyAvatar") +
          '" src="' +
          config.clientShareUserInfo.portrait +
          '" class="onekey-userinfo-avatar">' +
          '<div class="onekey-userinfo-name">' +
          config.clientShareUserInfo.displayname +
          "</div>" +
          '<div class="clientonekey-sumbit-button pass-button pass-button-submit" id="' +
          me.$getId("clientOnekeySumbit") +
          '">' +
          me.lang.clientOnekeySumbit +
          "</div>" +
          '<div class="onekey-change-password" id="' +
          me.$getId("clientOnekeyChangeLogin") +
          '">' +
          me.lang.clientOnekeyChangeLogin +
          "</div>";
        templateStr +=
          '<p id="' +
          me.$getId("clientShareWrapper") +
          '" class="pass-generalErrorWrapper">' +
          '<span id="' +
          me.$getId("clientShareError") +
          '" class="pass-generalError"></span>' +
          "</p>";
        templateStr += "</div>";
        return templateStr;
      },
      /**
       * 报错
       * @param {*} msg
       */
      _setClientShareError: function (msg) {
        var me = this;
        // bca-disable-next-line
        me.getElement("clientShareError") && (me.getElement("clientShareError").innerHTML = msg);
      },
      /**
       * 客户端互通一键登录事件
       */
      _setClientShareEvent: function () {
        var me = this;
        try {
          // 点击切换账密登录
          me.getElement("clientOnekeyChangeLogin") &&
            baidu(me.getElement("clientOnekeyChangeLogin")).on("click", function (evt) {
              var normalLogin = me._getTemplate();
              // 隐藏一键登录框
              me.$hide(baidu(me.getElement("clientonekey"))[0]);
              if (!me.getElement("form")) {
                // 如果是复合面板
                if (me.config.composeTemplate) {
                  // 添加两个外层box 横向布局
                  var composeRightNode = baidu("#componseRight").get(0);
                  composeRightNode.appendChild(baidu(normalLogin).get(0));
                } else {
                  var target = baidu(me.getElement());
                  target.get(0).appendChild(baidu(normalLogin).get(0));
                }
                // 设置输入框placeholder
                me.config.hasPlaceholder &&
                  me._getPlaceholder([
                    {
                      label: "userName",
                      placeholder: "userName",
                    },
                    {
                      label: "password",
                      placeholder: "password",
                    },
                  ]);
              } else {
                me.$show(baidu(me.getElement("form"))[0]);
              }
              if (me.config.makeText) {
                me.setMakeText(me.config.makeText);
              }
              me._setEvent();
              me.setEventFooterBar();
              me._setPasswordAgreeEvent();
              evt.preventDefault();
            });

          // 头像加载失败
          me.getElement("clientOnekeyAvatar") &&
            baidu(me.getElement("clientOnekeyAvatar")).on("error", function (evt) {
              evt.target.src = "https://himg.bdimg.com/sys/portrait/item/00000000.jpg";
            });

          // 点击直接登录
          var onekeySumbitFlag = false;
          me.getElement("clientOnekeySumbit") &&
            baidu(me.getElement("clientOnekeySumbit")).on("click", function (evt) {
              // 防止重复点击
              if (onekeySumbitFlag) {
                evt.preventDefault();
                return;
              }
              me._callBehaviorCallback({
                type: "client-onekey-sumbit",
              });
              onekeySumbitFlag = true;
              me.getElement("clientOnekeySumbit").style.opacity = "0.4";
              me.getElement("clientOnekeySumbit").innerText = me.lang.logining;
              var data = {
                tpl: me.config.product || "",
                authtoken: me.config.clientShareUserInfo.authToken || "",
                client: me.config.client || "pc",
                loginVersion: "v5",
                tt: new Date().getTime(),
                gid: me.guideRandom || "",
                u: me.config.u || "",
              };
              baidu.ajax({
                url: (passport.requestDomain || me._domain.https) + "/v3/login/webloginwithtoken",
                dataType: "jsonp",
                data: data,
                success: function (rsp) {
                  rsp = rsp || {};
                  rsp.loginType = "clientonekey";
                  if (rsp && rsp.errInfo && +rsp.errInfo.no === 0) {
                    try {
                      me.storageSupport(function () {
                        localStorage.setItem("passLoginType", "normal");
                      });
                    } catch (error) {
                      // console.log('_setClientShareEvent error', error);
                    }

                    var returnValue = me.fireEvent("loginSuccess", {
                      rsp: Object.assign({}, { type: "clientonekey" }, rsp),
                    });

                    if (!returnValue) {
                      return;
                    }
                    window.location.href = rsp.data.u;
                  } else {
                    onekeySumbitFlag = false;
                    me.getElement("clientOnekeySumbit").style.opacity = "1";
                    me.getElement("clientOnekeySumbit").innerText = me.lang.clientOnekeyChangeLogin;
                    var returnValue = me.fireEvent("loginError", {
                      rsp: Object.assign({}, { type: "clientonekey" }, rsp),
                    });
                    if (!returnValue) {
                      return;
                    }
                    me._setClientShareError((rsp.errInfo && rsp.errInfo.msg) || me.lang.networkError);
                  }
                },
                error: function () {
                  onekeySumbitFlag = false;
                  me.getElement("clientOnekeySumbit").style.opacity = "1";
                  me.getElement("clientOnekeySumbit").innerText = me.lang.clientOnekeyChangeLogin;
                  me._setClientShareError(me.lang.networkError);
                },
              });
              evt.preventDefault();
            });
        } catch (error) {
          // console.log('_setClientShareEvent error: ', error);
        }
      },
      /**
       * 初始化请求互通接口，判断是否能展示客户端互通一键登录
       */
      _initClientShare: function (callback) {
        var me = this;
        // 请求客户端互通接口
        if (me.config.isOpenClientShare && me.config.clientShareOpt) {
          // 获取客户端互通token
          window.passport.tryAllClientSharePorts(
            me.config.clientShareOpt,
            function (res) {
              if (res && +res.errno === 110000 && res.authToken) {
                me.config.clientShareIsLogin = true;
                me.config.clientShareUserInfo = res || {};
              } else {
                me.config.clientShareIsLogin = false;
              }
              typeof callback === "function" && callback();
            },
            function (error) {
              // 所有互通接口都失败，无登录态
              me.config.clientShareIsLogin = false;
              typeof callback === "function" && callback();
            },
          );
        } else {
          // 是否有互通登录态
          me.config.clientShareIsLogin = false;
          typeof callback === "function" && callback();
        }
      },
      _collectData: function () {
        var me = this;
        var dataObj = me.innerData["normal"];
        var fields = me.dataFiels;
        for (var i = 0, l = fields.length; i < l; i++) {
          dataObj[fields[i]] = me.getElement(fields[i]).value;
        }
        return dataObj;
      },
      _setRowDistance: function () {
        var me = this;
        var agreementWrapper = baidu(me.getElement("agreementWrapper")) && baidu(me.getElement("agreementWrapper"))[0];
        var agreementSmsWrapper =
          baidu(me.getElement("agreementSmsWrapper")) && baidu(me.getElement("agreementSmsWrapper"))[0];
        setTimeout(() => {
          var height = agreementWrapper && agreementWrapper.offsetHeight;
          var smsHeight = agreementSmsWrapper && agreementSmsWrapper.offsetHeight;
          if (height > 25 || smsHeight > 25) {
            baidu(me.getElement("submitWrapper")).addClass("customize-submit");
            baidu(me.getElement("smsSubmitWrapper")).addClass("customize-submit");
          }
        }, 0);
      },
      _restoreData: function (type) {
        var me = this;
        var dataObj = me.innerData[type ? type : "normal"];
        var fields = me.dataFiels;
        for (var i = 0, l = fields.length; i < l; i++) {
          me.getElement(fields[i]).value = dataObj[fields[i]] || "";
        }
        return dataObj;
      },

      _loadCssFileW: function (url, cb) {
        var me = this;
        window._loadedFilesW = window._loadedFilesW || {};
        if (!window._loadedFilesW[url]) {
          window._loadedFilesW[url] = true;
          var l = document.createElement("link");
          l.rel = "stylesheet";
          l.type = "text/css";
          l.href = passport.apiDomain.staticDomain + url;
          document.getElementsByTagName("head")[0].appendChild(l);
          if (l.readyState) {
            l.onreadystatechange = function () {
              if (l.readyState == "loaded" || l.readyState == "complete") {
                l.onreadystatechange = null;
                cb && cb();
              }
            };
          } else {
            l.onload = function () {
              cb && cb();
            };
          }
        } else {
          cb && cb();
        }
      },

      _insertScriptW: function (u, cb) {
        window._loadedFilesW = window._loadedFilesW || {};
        if (!window._loadedFilesW[u]) {
          window._loadedFilesW[u] = true;
          var d = document;
          var s = d.createElement("SCRIPT");
          s.type = "text/javascript";
          s.charset = "UTF-8";
          if (s.readyState) {
            s.onreadystatechange = function () {
              if (s.readyState == "loaded" || s.readyState == "complete") {
                s.onreadystatechange = null;
                cb && cb();
              }
            };
          } else {
            s.onload = function () {
              cb && cb();
            };
          }
          s.src = u;
          d.getElementsByTagName("head")[0].appendChild(s);
        } else {
          cb && cb();
        }
      },

      _authSiteW: function () {
        var me = this;
        var config = me.config;
        var targetCon = me.getPhoenixId("pass_phoenix_btn");
        if (config.authsitecssLoad) {
          me._loadCssFileW("/passApi/css/authsite_c01e2ff.css");
        }
        me._insertScriptW(me._domain.auto + "/phoenix/account/jsapi", function () {
          if (!window.baidu.phoenix) {
            return;
          }
          window.baidu.phoenix.require(
            config.authsiteLogin,
            baidu.extend(config.authsiteCfgLogin || {}, {
              // todo
              authsiteCurrentOpen: config.authsiteCurrentOpen,
              // authsiteCurrentOpen: true,
              tpl: config.product ? config.product : "",
              idc: config.idc ? config.idc : "",
              u: config.u,
              subpro: (config && config.subpro) || "",
              target: targetCon,
              appKey: config.appKey || "",
              staticPage: config.staticPage,
              canMessageChannel: me.config.canMessageChannel,
              html: {
                qzone: '<a class="phoenix-btn-item" href="#" data-title="qzone">QQ账号</a>',
                tsina: '<a class="phoenix-btn-item" href="#" data-title="tsina">新浪微博</a>',
                tqq: '<a class="phoenix-btn-item" href="#" data-title="tqq">腾讯微博</a>',
                qunar: '<a class="phoenix-btn-item" href="#" data-title="qunar">去哪儿</a>',
                weixin: '<a class="phoenix-btn-item" href="#" data-title="weixin">微信</a>',
                tianyi: '<a class="phoenix-btn-item" href="#" data-title="tianyi">天翼</a>',
                feifan: '<a class="phoenix-btn-item" href="#" data-title="feifan">飞凡</a>',
                dingtalk: '<a class="phoenix-btn-item" href="#" data-title="dingtalk">钉钉</a>',
                apple: '<a class="phoenix-btn-item" href="#" data-title="apple">apple</a>',
                facebook: '<a class="phoenix-btn-item" href="#" data-title="facebook">facebook</a>',
                twitter: '<a class="phoenix-btn-item" href="#" data-title="twitter">twitter</a>',
                google: '<a class="phoenix-btn-item" href="#" data-title="google">google</a>',
                google_bind: '<a class="phoenix-btn-item" href="#" data-title="google_bind">google</a>',
              },
              onAfterRender: function () {
                var $lists = baidu("#" + targetCon + " li");
                var preventDef = function (ele, $btnItem) {
                  // $btnItem 为 .phoenix-btn-item，用于取 data-title；ele 为父 li，覆盖完整点击范围
                  var $dataEle = $btnItem || ele;
                  ele.on("click", function (evt) {
                    if (!me._skipThirdClickLog) {
                      var autoStatisticObj = {
                        eventType: $dataEle.attr("data-title") + "Click",
                      };
                      me._logPass(me.urlData, autoStatisticObj);
                    }
                    me._skipThirdClickLog = false;
                    // thirdCheckAgreement=1且showAgreeCheckbox=1时校验当前tab协议是否勾选
                    if (+me.config.thirdCheckAgreement === 1 && me.config.showAgreeCheckbox) {
                      var currentTab = me.currentLoginType || "qrcode";
                      // compose模式下qrcode在左侧，右侧默认是账号密码tab
                      if (me.config.composeTemplate && currentTab === "qrcode") {
                        currentTab = "normal";
                      }
                      var checkboxId = "";
                      if (currentTab === "qrcode") {
                        checkboxId = me.$getId("qrcodeIsAgree");
                      } else if (currentTab === "sms") {
                        checkboxId = me.$getId("smsIsAgree");
                      } else {
                        checkboxId = me.$getId("isAgree");
                      }
                      var checkboxEle = document.getElementById(checkboxId);
                      if (checkboxEle && !checkboxEle.checked) {
                        evt.preventDefault();
                        evt.stopPropagation();
                        me._showProtocalConfirmDialog(function () {
                          checkboxEle.checked = true;
                          me._syncAgreeCheckboxState(currentTab, true);
                          // 跳过下次click的重复打点
                          me._skipThirdClickLog = true;
                          $dataEle[0] && $dataEle[0].click();
                        });
                        return;
                      }
                    }
                    evt.preventDefault();
                  });
                };
                for (var i = 0; i < $lists.length; i++) {
                  var $item = baidu(".phoenix-btn-item", $lists[i]);
                  $item.attr({ title: ($item[0] && $item[0].innerHTML) || "" });
                  // 拦截绑在li上，覆盖完整点击范围；打点从子元素取data-title
                  preventDef(baidu($lists[i]), $item);
                }
                me._logPass(me.urlData, {
                  eventType: "pc-authsite-show",
                });
              },
            }),
          );
        });
      },

      /**
       * @description 获取验证码
       * @name magic.passport.login#getVerifyCode
       * @function
       * @grammar magic.passport.login#getVerifyCode(codeString)
       * @param {String} codeString 验证码字符串
       */
      getVerifyCode: function (codeString) {
        var me = this;
        var data = {
          fr: "login",
          loginVersion: "v5",
          vcodetype: me.vcodetype || "",
        };
        me.getElement("verifyCode").value = "";
        me.$hide("verifyCodeSuccess");
        if (me.getElement("verifyCode_clearbtn")) {
          me.$hide("verifyCode_clearbtn");
        }
        me.getElement("verifyCodeImg").src = "";
        if (codeString && codeString.length) {
          me.$show("verifyCodeImgWrapper");
          me.getElement("verifyCodeImg").src = me.constant.VERIFYCODE_URL_PREFIX + codeString;
          me.getElement("codeString").value = codeString;
          var returnValue = me.fireEvent("renderVerifycode", {
            verifyStr: codeString,
            verifyCodeImg: me.constant.VERIFYCODE_URL_PREFIX + codeString,
          });
          if (!returnValue) {
            return;
          }
        } else {
          passport.data.getVerifyCodeStr(data).success(function (rsp) {
            if (rsp.errInfo.no == 0) {
              me.$show("verifyCodeImgWrapper");
              me.getElement("verifyCodeImg").src = me.constant.VERIFYCODE_URL_PREFIX + rsp.data.verifyStr;
              me.getElement("codeString").value = rsp.data.verifyStr;
              me.fireEvent("renderVerifycode", {
                verifyStr: rsp.data.verifyStr,
                verifyCodeImg: me.constant.VERIFYCODE_URL_PREFIX + rsp.data.verifyStr,
              });
            }
          });
        }
        if (me.getElement("verifyCode_placeholder")) {
          setTimeout(function () {
            me.$show("verifyCode_placeholder");
          }, 200);
        }
      },

      _getWDom: {
        parent: function (dom) {
          return dom.parentNode || dom.parentElement;
        },
        next: function (dom) {
          do {
            dom = dom.nextSibling;
          } while (dom && dom.nodeType !== 1);
          return dom;
        },
        prev: function (dom) {
          do {
            dom = dom.previousSibling;
          } while (dom && dom.nodeType !== 1);
          return dom;
        },
      },

      /**
       * @description 清除验证码
       * @name magic.passport.login#clearVerifyCode
       * @function
       * @grammar magic.passport.login#clearVerifyCode()
       */
      clearVerifyCode: function () {
        var me = this;
        me.$hide("verifyCodeImgWrapper");
        me.getElement("codeString").value = "";
      },
      getPhoenixId: function (field) {
        if (this.config.uniqueId) {
          return this.$getId(field);
        } else {
          var phoenixEles = {
            pass_phoenix_login: "pass-phoenix-login",
            pass_phoenix_list_login: "pass-phoenix-list-login",
            pass_phoenix_btn: "pass_phoenix_btn",
          };
          return phoenixEles[field];
        }
      },
      getPhoenixElement: function (field) {
        if (this.getElement(field)) {
          return this.getElement(field);
        } else {
          return document.getElementById(this.getPhoenixId(field));
        }
      },
      _getTemplateOther: function () {
        var html = [];
        var me = this;
        var length = 0;
        if (me.config.authsiteLogin) {
          length = me.config.authsiteLogin.length;
        }
        html.push(
          '<div id="' +
            me.getPhoenixId("pass_phoenix_login") +
            '" class="tang-pass-login-phoenix">' +
            '<div id="' +
            me.getPhoenixId("pass_phoenix_list_login") +
            '" class="pass-phoenix-list clearfix">' +
            (me.config.authsiteLogin
              ? '<div class="pass-phoenix-btn clearfix" id="' + me.getPhoenixId("pass_phoenix_btn") + '"></div>'
              : "") +
            "</div>" +
            '<div class="clear"></div>' +
            "</div>",
        );
        return html.join("");
      },
      getTemplateFooterBar: function () {
        var me = this;
        var footerBarHtml = [];
        var passLoginType = "";
        me.storageSupport(function () {
          passLoginType = localStorage.getItem("passLoginType");
        });

        // footer是否展示扫码登录 （默认扫码line 76 所以这里默认是底部不展示扫码登录）
        var showQrcodeFooter =
          (+me.config.userPwdLogin === 1 || (passLoginType === "sms" && me.smsConfig) || passLoginType === "normal") &&
          +me.config.qrcodeLogin !== 1;

        footerBarHtml.push(
          '<div class="tang-pass-footerBar">' +
            // todo 需要去掉这里逻辑
            // + (me.config.qrcode === 3
            //     ? ('<p class="tang-pass-footerBarQrcode pass-link" title="'
            //         + me.lang.qrcodeBtn
            //         + '" data-type="qrcode" id="'
            //         + me.$getId('footerQrcodeBtn')
            //         + '"'
            //         + (!showQrcodeFooter ? ' style="display:none;"' : '')
            //         + '>' + me.lang.qrcodeBtn + '</p>') : '')
            // + '<p class="tang-pass-footerBarULogin pass-link" title="'
            // + me.lang.footerBackToLogin
            // + '" data-type="normal" id="'
            // + me.$getId('footerULoginBtn')
            // + '"' + (showQrcodeFooter ? ' style="display:none;"' : '')
            // + '>' + me.lang.footerBackToLogin + '</p>'
            (me.config.authsiteLogin.length > 0
              ? '<div class="tang-pass-footerBarPhoenix">' +
                // + '<span class="tang-pass-footerBarPhoenixSplit"></span>'
                '<div class="tang-pass-footerBarPhoenixItem" id="' +
                me.$getId("PhoenixItem") +
                '"></div></div>'
              : "") +
            // 切换海外登录入口
            (me.config.isOpenOverseasLogin
              ? '<div class="switch-overseas-btn ' +
                (me.config.composeTemplate ? "s-overseas-btn-compose" : "") +
                '" id="' +
                me.$getId("switchToOverseasBtn") +
                '">' +
                (me.config.overseasLoginOpt.customText.switchToOverseas || me.lang.switchOverseasLogin) +
                "</div>"
              : "") +
            (me.config.overseas && me.config.isOpenOverseasLogin
              ? ""
              : '<a class="pass-reglink pass-link" id="' +
                me.$getId("regLink") +
                '"' +
                (me.config.hasRegUrl ? "" : 'style="display:none"') +
                ">" +
                me.lang.register +
                "</a>") +
            "</div>",
        );
        return footerBarHtml.join("");
      },
      setLoginTypeTabBar: function () {
        var me = this;
        var qrcodeBtn = me.$getId("changeQrCodeItem");
        var uLoginBtn = me.$getId("changePwdCodeItem");
        var smsBtn = me.$getId("changeSmsCodeItem");
        var downloadBtn = me.$getId("QrcodeDownload");
        baidu("#" + qrcodeBtn).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._changeLoginType("qrcode");
          me._logPass(me.urlData, {
            eventType: "user-qrLogin-click",
          });
          me._callBehaviorCallback({
            type: "user-qrLogin-click",
          });
        });
        baidu("#" + uLoginBtn).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._logPass(me.urlData, {
            eventType: "user-pwdLogin-click",
          });
          me._callBehaviorCallback({
            type: "user-pwdLogin-click",
          });
          me._changeLoginType("normal");
        });
        baidu("#" + smsBtn).on("click", function () {
          if (/msie 6/i.test(navigator.userAgent) || /msie 7/i.test(navigator.userAgent)) {
            setTimeout(function () {
              me._changeLoginType("sms");
            }, 0);
          } else {
            me._changeLoginType("sms");
          }
          me._logPass(me.urlData, {
            eventType: "user-smsLogin-click",
          });
          me._callBehaviorCallback({
            type: "user-smsLogin-click",
          });
        });
        // 在这里把点击下载打点初始化一下
        baidu("#" + downloadBtn).on("click", function () {
          var autoStatisticObj1 = {
            eventType: "toDownload",
          };
          me._logPass(me.urlData, autoStatisticObj1);
        });
      },
      _changeTabButtonActivStatus: function (type) {
        // todo dadian
        var me = this;
        me.config.currentTab = type;
        var qrcodeBtn = me.$getId("changeQrCodeItem");
        var uLoginBtn = me.$getId("changePwdCodeItem");
        var smsBtn = me.$getId("changeSmsCodeItem");
        if (me.config.composeTemplate && type === "qrcode") {
          baidu("#" + uLoginBtn).addClass("activ");
          baidu("#" + qrcodeBtn).removeClass("activ");
          baidu("#" + smsBtn).removeClass("activ");
          me._logPass(me.urlData, {
            eventType: "pc-pwdLogin-show",
          });
        } else {
          if (type === "sms") {
            baidu("#" + smsBtn).addClass("activ");
            baidu("#" + qrcodeBtn).removeClass("activ");
            baidu("#" + uLoginBtn).removeClass("activ");
          } else if (type === "qrcode") {
            baidu("#" + qrcodeBtn).addClass("activ");
            baidu("#" + uLoginBtn).removeClass("activ");
            baidu("#" + smsBtn).removeClass("activ");
          } else {
            baidu("#" + uLoginBtn).addClass("activ");
            baidu("#" + qrcodeBtn).removeClass("activ");
            baidu("#" + smsBtn).removeClass("activ");
          }
        }
      },
      setEventFooterBar: function () {
        var me = this;
        var qrcodeBtn = me.$getId("footerQrcodeBtn");
        var uLoginBtn = me.$getId("footerULoginBtn");
        var regLink = me.$getId("regLink");
        var fgtpwdLink = me.$getId("fgtpwdLink");

        baidu("#" + qrcodeBtn).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._changeLoginType("qrcode");
          baidu("#" + qrcodeBtn).hide();
          baidu("#" + uLoginBtn).show();
        });
        baidu("#" + uLoginBtn).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }

          me._changeLoginType("normal");
          baidu("#" + uLoginBtn).hide();
          baidu("#" + qrcodeBtn).show();
        });
        baidu("#" + regLink).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._logPass(me.urlData, {
            eventType: "pc-register-click",
          });
          me._callBehaviorCallback({
            type: "pc-register-click",
          });
          window.open(me.constant.REG_URL, "_blank");
        });
        baidu("#" + fgtpwdLink).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._logPass(me.urlData, {
            eventType: "pc-forget-click",
          });
          me._callBehaviorCallback({
            type: "pc-forget-click",
          });
          if (me.isEnVersion) {
            me._ownerDialog && me._ownerDialog.hide("unHide");
            me._insertScriptW(me.constant.RESETPASSWORD_URL, function () {
              if (window.passResetpwdSdk) {
                window.passResetpwdSdk.passResetpwdInit({
                  tpl: me.config.product || "",
                  lang: "en",
                  themeColor: "#17171A",
                  closeFn: function () {
                    me._ownerDialog && me._ownerDialog.show();
                  },
                  backFn: function () {
                    me._ownerDialog && me._ownerDialog.show();
                  },
                });
              }
            });
          } else {
            window.open(me.constant.GET_PASSWORD_URL, "_blank");
          }
        });
        me.setOverseasLogin();
      },

      setOverseasLogin: function () {
        var me = this;
        var switchToOverseasBtn = me.$getId("switchToOverseasBtn");
        baidu("#" + switchToOverseasBtn).on("click", function (evt) {
          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }

          me._ownerDialog && me._ownerDialog.hide();
          if (window.passOverseasLoginSdk) {
            const { passOverseasLoginInit } = window.passOverseasLoginSdk;
            passOverseasLoginInit({
              lang: me.config.lang,
              tpl: me.config.overseasLoginOpt.tpl || "",
              u: me.config.overseasLoginOpt.u || "",
              customText: me.config.overseasLoginOpt.customText || {},
              themeColor: me.config.overseasLoginOpt.themeColor || "#4E6EF2",
              // 自定义协议
              agreements: me.config.overseasLoginOpt.agreements || [],
              // 轮播图最多传三张
              bannerImages: me.config.overseasLoginOpt.bannerImages || [],
              // 配置需要显示的三方登录方式, 不传则默认只显示谷歌登录
              thirdLoginTypes: me.config.overseasLoginOpt.thirdLoginTypes || [],
              // 关闭登录框回调
              onClose: me.config.overseasLoginOpt.onClose || function () {},
              // 切换到国内版本回调
              onSwitchToChinaVersion: function () {
                me._ownerDialog && me._ownerDialog.show();
                me.config.overseasLoginOpt.onSwitchToChinaVersion &&
                  me.config.overseasLoginOpt.onSwitchToChinaVersion();
              },
              // 渲染完成回调
              renderCallback: me.config.overseasLoginOpt.renderCallback || function () {},
            });
          }
        });
      },

      /**
       * 微信小程序扫码登录二维码模版
       * @returns {String}
       */
      _getTemplateWxMiniQrcode: function () {
        var me = this;
        var qrcodeHtml = [];
        var isBlock = me.config.composeTemplate && me.config.isOpenWxMiniQrcodeLogin ? "block" : "none";
        qrcodeHtml.push(
          '<div id="' +
            me.$getId("wxQrcode") +
            '" class="clearfix tang-pass-qrcode tang-pass-login" style="display:' +
            isBlock +
            ';">',
        );
        qrcodeHtml.push('<div class="tang-pass-qrcode-content" id="' + me.$getId("wxQrcodeContent") + '">');
        qrcodeHtml.push('<div class="tang-pass-qrcode-init">');
        qrcodeHtml.push(
          '<div class="Qrcode-status-con tang-pass-qrcode-imgWrapper wxQrcode-status-con" id="' +
            me.$getId("wxQrcodeMain") +
            '">' +
            '<img class="tang-pass-qrcode-img" id="' +
            me.$getId("wxQrcodeImg") +
            '" src="' +
            passport.apiDomain.staticDomain +
            '/passApi/img/loading.gif"/>' +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-success wxQrcode-status-con" id="' +
            me.$getId("wxQrcodeSuccess") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            "<p>" +
            me.lang.wxQrcodeSuccessTip +
            "</p>" +
            '<p class="Qrcode-status-msg">' +
            me.lang.wxMiniSureLogin +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-error wxQrcode-status-con" id="' +
            me.$getId("wxQrcodeError") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            '<p class="refresh-title">' +
            me.lang.QrcodeErrorTip +
            "</p>" +
            '<p class="Qrcode-refresh-btn" id="' +
            me.$getId("wxQrcodeErrorfreshBtn") +
            '">' +
            me.lang.QrcodeRefreshBtn +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-error wxQrcode-status-con" id="' +
            me.$getId("wxQrcodeRisk") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            '<p class="refresh-title">' +
            me.lang.QrcodeRiskTip +
            "</p>" +
            '<p class="Qrcode-status-msg">' +
            me.lang.QrcodeErrorMsg +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-refresh wxQrcode-status-con" id="' +
            me.$getId("wxQrcodeRefresh") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            '<p class="refresh-title refresh-timeout">' +
            me.lang.QrcodeRefreshTip +
            "</p>" +
            '<p class="refresh-title refresh-loadout">' +
            me.lang.QrcodeLoadTip +
            "</p>" +
            '<p class="refresh-title refresh-cancel">' +
            me.lang.qrcodeCancelTip +
            "</p>" +
            '<p class="Qrcode-refresh-btn" id="' +
            me.$getId("wxQrcodeRefreshBtn") +
            '">' +
            me.lang.QrcodeRefreshBtn +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push("</div>");
        qrcodeHtml.push("</div>");
        qrcodeHtml.push(
          '<p id="' +
            me.$getId("wxQrcodeTitle") +
            '" class="tang-pass-qrcode-title">' +
            me.lang.useWxMiniLogin +
            "</p>",
        );
        qrcodeHtml.push(
          '<p class="tang-pass-qrcode-changelogin" id="' +
            me.$getId("changeBdQrLoginTitle") +
            '">' +
            me.lang.click +
            '<span class="pass-qrcode-changelogin-clicktext" id="' +
            me.$getId("changeBdQrAppLogin") +
            '">' +
            me.lang.changeBaiduQrCode +
            "</span>" +
            "</p>",
        );
        qrcodeHtml.push("</div>");
        return qrcodeHtml.join("");
      },
      /**
       * 微信二维码绑定相关事件
       */
      _setWxMiniQrcodeEvent: function () {
        var me = this;
        // 刷新二维码
        me.getElement("wxQrcodeErrorfreshBtn") &&
          baidu(me.getElement("wxQrcodeErrorfreshBtn")).on("click", function () {
            me._setWxMiniQrChannel();
          });
        me.getElement("wxQrcodeRefreshBtn") &&
          baidu(me.getElement("wxQrcodeRefreshBtn")).on("click", function () {
            me._setWxMiniQrChannel();
          });
        // 切换为百度二维码
        me.getElement("changeBdQrAppLogin") &&
          baidu(me.getElement("changeBdQrAppLogin")).on("click", function () {
            // 停止微信小程序轮询
            me._stopWxMiniQrChannel();
            // 如果存在百度扫码，则切换并开启轮询，否则添加百度二维码视图
            if (me.getElement("qrcode")) {
              me.$hide(me.getElement("wxQrcode"));
              me.$show(me.getElement("qrcode"));
              me._setChannel();
            } else {
              me.$hide(me.getElement("wxQrcode"));
              var composeLeftNode = document.getElementById("componseLeft");
              var qrcodeTemplate = me._getTemplateQrcode();
              composeLeftNode.appendChild(baidu(qrcodeTemplate).get(0));
              me._actionQrcode();
              me.setqrcodeEvent();
            }
          });
      },
      /**
       * 获取扫码开启轮询sign/channel id透传给业务方
       */
      _setWxMiniQrChannel: function () {
        var me = this;
        var container = me.getElement("wxQrcodeContent");
        try {
          var data = {
            client: "pc",
            clientfrom: "pc",
            qrloginfrom: "pc",
            tt: new Date().getTime(),
            tpl: me.config.product || "",
          };

          // 存储微信小程序扫码轮询信息
          passport.wxMiniQrcodeData = passport.wxMiniQrcodeData || {};
          baidu(".Qrcode-status-con", container).hide();
          me.$show(me.getElement("wxQrcodeMain"));
          me.$show(me.getElement("wxQrcodeTitle"));
          me.$show(me.getElement("changeBdQrLoginTitle"));
          if (me.getElement("wxQrcodeImg")) {
            me.getElement("wxQrcodeImg").src = passport.apiDomain.staticDomain + "/passApi/img/loading.gif";
          }
          // 结束上一个轮询
          me._stopWxMiniQrChannel();
          passport.wxMiniQrcodeData.loadQrcode = setTimeout(function () {
            if (me.config.qrcode_style) {
              baidu(".Qrcode-status-con", container).hide();
              baidu(".refresh-title", container).hide();
              baidu(".refresh-loadout", container).show();
              me.$hide(me.getElement("wxQrcodeTitle"));
              me.$hide(me.getElement("changeBdQrLoginTitle"));
              me.$show(me.getElement("wxQrcodeRefresh"));
            }
          }, 10 * 1000);
          var url = me._domain.https + "/v3/api/mini/genchannelid";
          passport.sendAjax({
            url: url,
            method: "post",
            data,
            success: function (args) {
              clearTimeout(passport.wxMiniQrcodeData.loadQrcode);
              // 操作频繁
              if (+args.code === 50000) {
                baidu(".Qrcode-status-con").hide();
                me.$hide(me.getElement("wxQrcodeTitle"));
                me.$hide(me.getElement("changeBdQrLoginTitle"));
                me.$show(me.getElement("wxQrcodeRisk"));
                return;
              }
              passport.wxMiniQrcodeData.sign = args.data.channelID;
              // 开启扫码轮询
              passport.wxMiniQrcodeData.isOpenWxPolling = true;
              // 回调业务方，业务方生成微信小程序二维码，参数id为二维码元素id，sign为channel id
              me.config.updateWxMiniQrcode &&
                typeof me.config.updateWxMiniQrcode === "function" &&
                me.config.updateWxMiniQrcode(me.$getId("wxQrcodeImg"), args.data.channelID, args.data.imgurl);

              setTimeout(function () {
                me._createWxMiniQrChannel(passport.wxMiniQrcodeData.sign);
              }, 300);
              // 10分钟后微信小程序二维码过期，需要手动刷新
              passport.wxMiniQrcodeData.timer = setTimeout(
                function () {
                  me._stopWxMiniQrChannel();
                  baidu(".Qrcode-status-con", container).hide();
                  baidu(".refresh-title", container).hide();
                  baidu(".refresh-timeout", container).show();
                  me.$hide(me.getElement("wxQrcodeTitle"));
                  me.$hide(me.getElement("changeBdQrLoginTitle"));
                  me.$show(me.getElement("wxQrcodeRefresh"));
                  return;
                },
                10 * 59 * 1000,
              );
            },
            error: function () {
              baidu(".Qrcode-status-con", container).hide();
              me.$hide(me.getElement("wxQrcodeTitle"));
              me.$hide(me.getElement("changeBdQrLoginTitle"));
              me.$show(me.getElement("wxQrcodeError"));
            },
          });
        } catch (error) {
          baidu(".Qrcode-status-con", container).hide();
          me.$hide(me.getElement("wxQrcodeTitle"));
          me.$hide(me.getElement("changeBdQrLoginTitle"));
          me.$show(me.getElement("wxQrcodeError"));
        }
      },
      /**
       * 微信小程序扫码停止轮询
       */
      _stopWxMiniQrChannel: function () {
        passport.wxMiniQrcodeData = passport.wxMiniQrcodeData || {};
        passport.wxMiniQrcodeData.sign = "";
        // 关闭微信扫码轮询
        passport.wxMiniQrcodeData.isOpenWxPolling = false;
        clearTimeout(passport.wxMiniQrcodeData && passport.wxMiniQrcodeData.timer);
        clearTimeout(passport.wxMiniQrcodeData && passport.wxMiniQrcodeData.loadQrcode);
      },
      /**
       * 微信小程序扫码开始轮询
       * @param {*} sign
       * @returns
       */
      _createWxMiniQrChannel: function (sign) {
        var me = this;
        var qrcodeSign = sign;
        var container = me.getElement("wxQrcodeContent");
        var qrcodeInit = baidu(".tang-pass-qrcode-init", container).get(0);
        var qrcodeImg = baidu(".tang-pass-qrcode-img", qrcodeInit).get(0);
        var data = {
          apiver: "v3",
          tt: new Date().getTime(),
        };
        passport.wxMiniQrcodeData = passport.wxMiniQrcodeData || {};
        if (!passport.wxMiniQrcodeData.sign) {
          return false;
        }
        baidu.ajax({
          url:
            me._domain.https +
            "/channel/unicast?channel_id=" +
            passport.wxMiniQrcodeData.sign +
            "&gid=" +
            (me.guideRandom || "") +
            "&tpl=" +
            me.config.product +
            "&_sdkFrom=1",
          dataType: "jsonp",
          data: data,
          success: function (d) {
            if (d.channel_v) {
              try {
                d.channel_v = JSON.parse(d.channel_v);
              } catch (e) {
                d.channel_v = {};
              }
            } else {
              d.channel_v = {};
            }
            if (d.channel_v.u) {
              d.channel_v.u = decodeURIComponent(d.channel_v.u);
            }
            // 扫码成功后确认授权
            if (+d.errno === 0 && +d.channel_v.status === 0) {
              clearTimeout(passport.wxMiniQrcodeData.timer);
              var data = {
                sign: d.channel_v.v,
                u: encodeURIComponent(d.channel_v.u || me.config.u),
                loginVersion: "v5",
                tpl: me.config.product ? me.config.product : "",
                client: "pc",
                clientfrom: "pc",
                qrcodefrom: "pc",
              };
              var url = me._domain.https + "/v3/api/mini/qrlogin";
              passport.sendAjax({
                url: url,
                method: "post",
                data,
                success: function (args) {
                  var rsp = JSON.parse(args);
                  rsp.loginType = "wxminiqrcode";

                  if (rsp.errInfo && +rsp.errInfo.no === 110000) {
                    var returnValue = me.fireEvent("loginSuccess", {
                      rsp: Object.assign({}, { type: "qrCode" }, rsp),
                    });
                    if (!returnValue) {
                      return;
                    }
                    window.location.href = rsp.data.u;
                  } else {
                    // 二维码失效
                    // 310004, 310005轮询接口bduss入参失效
                    // 400021，400022 小程序端登录态失效
                    clearTimeout(passport.wxMiniQrcodeData.timer);
                    if (me.config.qrcode_style) {
                      baidu(".Qrcode-status-con", container).hide();
                      baidu(".refresh-title", container).hide();
                      baidu(".refresh-timeout", container).show();
                      me.$hide(me.getElement("wxQrcodeTitle"));
                      me.$hide(me.getElement("changeBdQrLoginTitle"));
                      me.$show(me.getElement("wxQrcodeRefresh"));
                    } else {
                      if (qrcodeImg) {
                        qrcodeImg.src = passport.apiDomain.staticDomain + "/passApi/img/loading.gif";
                      }
                      me._setWxMiniQrChannel();
                    }
                    me.fireEvent("loginError", {
                      rsp: Object.assign({}, { type: "qrCode" }, rsp),
                    });
                    me._logPass(me.urlData, {
                      eventType: "pc-qrLogin-submit-error",
                      errno: rsp.errInfo.no,
                    });
                  }
                },
              });
            } else {
              // 扫码成功
              if (d.errno == 0 && d.channel_v.status == "1") {
                if (me.config.qrcode_style) {
                  baidu(".Qrcode-status-con", container).hide();
                  me.$hide(me.getElement("wxQrcodeTitle"));
                  me.$hide(me.getElement("changeBdQrLoginTitle"));
                  me.$show(me.getElement("wxQrcodeSuccess"));
                }
                if (passport.wxMiniQrcodeData.sign && qrcodeSign == passport.wxMiniQrcodeData.sign) {
                  me._createWxMiniQrChannel(qrcodeSign);
                }
              } else if (d.errno == 0 && d.channel_v.status == "2") {
                // 取消扫码
                clearTimeout(passport.wxMiniQrcodeData.timer);
                var returnValue = me.fireEvent("loginQrcodeCancel");
                if (!returnValue) {
                  return;
                }
                if (me.config.qrcode_style) {
                  baidu(".Qrcode-status-con", container).hide();
                  baidu(".refresh-title", container).hide();
                  baidu(".refresh-timeout", container).show();
                  me.$hide(me.getElement("wxQrcodeTitle"));
                  me.$hide(me.getElement("changeBdQrLoginTitle"));
                  me.$show(me.getElement("wxQrcodeRefresh"));
                } else {
                  if (qrcodeImg) {
                    qrcodeImg.src = passport.apiDomain.staticDomain + "/passApi/img/loading.gif";
                  }
                  me._setWxMiniQrChannel();
                }
              } else if (passport.wxMiniQrcodeData.sign && qrcodeSign == passport.wxMiniQrcodeData.sign) {
                me._createWxMiniQrChannel(qrcodeSign);
              }
            }
          },
          error: function () {
            if (me.config.qrcode_style) {
              baidu(".Qrcode-status-con", container).hide();
              me.$hide(me.getElement("wxQrcodeTitle"));
              me.$hide(me.getElement("changeBdQrLoginTitle"));
              me.$show(me.getElement("wxQrcodeError"));
            }
          },
        });
      },
      _getTemplateQrcode: function () {
        var me = this;
        var qrcodeHtml = [];
        var guideMapImg =
          me.config.guideMapImg || "https://ppui-static-pc.cdn.bcebos.com/passApi/img/qrcodeLoginGuide1.png";
        var appHref = (me.config.qrcodeCfg && me.config.qrcodeCfg.appHref) || "https://mo.baidu.com/wuxian/?from=pass";
        var appName = (me.config.qrcodeCfg && me.config.qrcodeCfg.appName) || "百度APP";
        var appIcon =
          (me.config.qrcodeCfg && me.config.qrcodeCfg.appIcon) ||
          "https://ppui-static-pc.cdn.bcebos.com/passApi/img/baidu-download-icon.png";
        var downLoadQrcode =
          (me.config.qrcodeCfg && me.config.qrcodeCfg.downLoadQrcode) ||
          "https://b.bdstatic.com/searchbox/icms/searchbox/img/code_home.png";
        var disableJump = (me.config.qrcodeCfg && me.config.qrcodeCfg.disableJump) || false;
        var isBlock = me.config.composeTemplate ? "block" : "none";
        qrcodeHtml.push(
          '<div id="' +
            me.$getId("qrcode") +
            '" class="clearfix tang-pass-qrcode tang-pass-login" style="display:' +
            isBlock +
            ';">',
        );
        qrcodeHtml.push('<div class="tang-pass-qrcode-content" id="' + me.$getId("qrcodeContent") + '">');
        qrcodeHtml.push('<div class="tang-pass-qrcode-init">');
        qrcodeHtml.push(
          '<div class="Qrcode-status-con tang-pass-qrcode-imgWrapper" id="' +
            me.$getId("QrcodeMain") +
            '">' +
            '<img class="tang-pass-qrcode-img" src="' +
            passport.apiDomain.staticDomain +
            '/passApi/img/loading.gif"/>' +
            (me.config.qrcode_animation && !me.config.isOpenWxMiniQrcodeLogin
              ? // 【重要】Qrcode-status-animation 样式名称不能改，更改影响网盘自定义样式
                '<p class="Qrcode-status-animation' +
                (+me.passLowerIE === 8 ? " Qrcode-status-animation-hackIE8" : "") +
                '" id="' +
                me.$getId("QrcodeAnimation") +
                '" style="background: url(' +
                guideMapImg +
                ') no-repeat center center / 134px 160px"></p>'
              : "") +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-success" id="' +
            me.$getId("QrcodeSuccess") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            "<p>" +
            me.lang.QrcodeSuccessTip +
            "</p>" +
            '<p class="Qrcode-status-msg">' +
            me.lang.QrcodeSuccessMsg +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-error" id="' +
            me.$getId("QrcodeError") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            "<p>" +
            me.lang.QrcodeErrorTip +
            "</p>" +
            '<p class="Qrcode-refresh-btn" id="' +
            me.$getId("QrcodeErrorfreshBtn") +
            '">' +
            me.lang.QrcodeRefreshBtn +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-error" id="' +
            me.$getId("QrcodeRisk") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            "<p>" +
            me.lang.QrcodeRiskTip +
            "</p>" +
            '<p class="Qrcode-status-msg">' +
            me.lang.QrcodeErrorMsg +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push(
          '<div class="Qrcode-status-con Qrcode-status-refresh" id="' +
            me.$getId("QrcodeRefresh") +
            '">' +
            '<p class="Qrcode-status-icon"></p>' +
            '<p class="refresh-title refresh-timeout">' +
            me.lang.QrcodeRefreshTip +
            "</p>" +
            '<p class="refresh-title refresh-loadout">' +
            me.lang.QrcodeLoadTip +
            "</p>" +
            '<p class="Qrcode-refresh-btn" id="' +
            me.$getId("QrcodeRefreshBtn") +
            '">' +
            me.lang.QrcodeRefreshBtn +
            "</p>" +
            "</div>",
        );
        qrcodeHtml.push("</div>");
        qrcodeHtml.push("</div>");
        // 如果开启了微信小程序登录，则显示切换登录方式按钮
        if (me.config.isOpenWxMiniQrcodeLogin) {
          qrcodeHtml.push(
            '<p id="' +
              me.$getId("QrcodeTitle") +
              '" class="tang-pass-qrcode-title">' +
              me.lang.qrcodeTitle +
              appName +
              me.lang.qrcodeText +
              "</p>",
          );
          qrcodeHtml.push(
            '<p class="tang-pass-qrcode-changelogin" id="' +
              me.$getId("changeWxQrLoginTitle") +
              '">' +
              me.lang.click +
              '<span class="pass-qrcode-changelogin-clicktext" id="' +
              me.$getId("changeWxQrAppLogin") +
              '">' +
              me.lang.changeWxQrCode +
              "</span>" +
              "</p>",
          );
        } else {
          if (me.config.qrcodeCfg && me.config.qrcodeCfg.appTitle) {
            qrcodeHtml.push(
              '<p id="' +
                me.$getId("QrcodeTitle") +
                '" class="tang-pass-qrcode-title">' +
                me.config.qrcodeCfg.appTitle +
                "</p>",
            );
          } else {
            // qrcode title 请使用百度APP扫码登录
            // qrcodeHtml.push('<p class="tang-pass-qrcode-title">' + me.lang.qrcodeTitle
            // + (me.config.qrcodeCfg ? '<a class="pass-link" href="' + appHref + '">' + appName
            // + '</a>' : me.lang.qrcodeHref) + me.lang.qrcodeText + '</p>');
            qrcodeHtml.push(
              '<p id="' +
                me.$getId("QrcodeTitle") +
                '" class="tang-pass-qrcode-title">' +
                me.lang.qrcodeTitle +
                appName +
                me.lang.qrcodeText +
                "</p>",
            );
            // 带有icon下载按钮
            qrcodeHtml.push(
              '<div id="' +
                me.$getId("QrcodeDownload") +
                '" class="pass-qrcode-download">' +
                '<div class="qrcode-hover"><div class="qrcode-download-link-box">' +
                '<img class="qrcode-download-link" src="' +
                downLoadQrcode +
                '">' +
                "</div><p>扫描二维码下载" +
                appName +
                "</p></div>" +
                '<a class="pass-link" href="' +
                (disableJump ? "javascript:;" : appHref) +
                '" target="_blank">' +
                '<img class="pass-qrcode-icon" src="' +
                appIcon +
                '" alt="">' +
                "下载" +
                appName +
                "</a></div>",
            );
          }
          if (me.config.qrcodeCfg && me.config.qrcodeCfg.appSubTitle) {
            qrcodeHtml.push('<p class="tang-pass-qrcode-subtitle">' + me.config.qrcodeCfg.appSubTitle + "</p>");
          }
        }
        // qrcodeHtml.push('<ul class="tang-pass-qrcode-ullist">'
        //     + '<li class="tang-pass-qrcode-list">'
        //     + '<span class="tang-pass-qrcode-list-aq"></span>'
        //     + '<span>' + me.lang.qrcodeListaq + '</span></li>'
        //     + '<li class="tang-pass-qrcode-list">'
        //     + '<span class="tang-pass-qrcode-list-gx"></span>'
        //     + '<span>' + me.lang.qrcodeListgx + '</span></li>'
        //     + '<li class="tang-pass-qrcode-list">'
        //     + '<span class="tang-pass-qrcode-list-bj"></span>'
        //     + '<span>' + me.lang.qrcodeListbj + '</span></li>'
        //     + '</ul>');
        // 扫码tab协议勾选框（仅非compose模式 + showAgreeCheckbox=1 + thirdCheckAgreement=1时展示）
        if (+me.config.thirdCheckAgreement === 1 && me.config.showAgreeCheckbox && !me.config.composeTemplate) {
          qrcodeHtml.push(
            '<p class="pass-form-item"><span id="' +
              me.$getId("agreementQrcodeWrapper") +
              '" class="tang-pass-sms-agreement pass-link">' +
              '<input name="qrcodeIsAgree" id="' +
              me.$getId("qrcodeIsAgree") +
              '" type="checkbox" class="pass-checkbox-input pass-checkbox-isAgree" autocomplete="off" />' +
              '<label for="' +
              me.$getId("qrcodeIsAgree") +
              '" style="padding-left: 5px;">' +
              me.lang.agree +
              "</label>" +
              '<a target="_blank" href="' +
              me.constant.PROTOCAL_URL +
              '" style="display: inline-block;margin-bottom: 6px">' +
              (me.config.protocalText || me.lang.baiduUserProtocal) +
              "</a> " +
              (me.customizeAnd ? "、" : me.lang.and) +
              '<a target="_blank" href="' +
              me.constant.PERSONAL_PROTOCAL_URL +
              '">' +
              (me.config.userpersonalText || me.lang.baiduUserPersonal) +
              "</a>" +
              (me.config.showChildAgree
                ? '、<a target="_blank" href="' +
                  me.constant.CHILD_AGREEMENT_URL +
                  '">' +
                  me.lang.childAgreement +
                  "</a> "
                : "") +
              "</span></p>",
          );
        }
        qrcodeHtml.push("</div>");
        return qrcodeHtml.join("");
      },
      // 切换成二维码登录
      _setEventQrcode: function () {
        var me = this;
        // var passLoginType = '';
        // me.storageSupport(function () {
        //     passLoginType = localStorage.getItem('passLoginType');
        // });
        // （如果产品线没配置账密 && 上次登陆方式是二维码） || （没上次登录方式 && pass默认的二维码登录方式） || 产品线配置了二维码
        // if (passLoginType === 'qrcode'
        //     || (!passLoginType && me.config.qrcode === 3)
        //     || me.config.qrcodeLogin === 1) {
        if (/msie 6/i.test(navigator.userAgent)) {
          setTimeout(function () {
            me._changeLoginType("qrcode");
          }, 0);
        } else {
          me._changeLoginType("qrcode");
        }
        // }
      },

      // 短信故障源-获取channel
      getChannelForUnReceiveSms: function (preUpSms) {
        var me = this;
        var phone = document.getElementById(me.$getId("smsPhone"));
        me.config.unReceiveSmsCode = "";
        // 先结束扫码登录的轮询，再开始上行短信登录轮询
        me._stopChannel();
        baidu.ajax({
          url: (passport.requestDomain || me._domain.https) + "/v3/login/getChannel",
          dataType: "jsonp",
          data: {
            from: "up_sms_login",
            data: me._SBCtoDBC(phone.value),
          },
          success: function (rsp) {
            if (+rsp.errno === 0) {
              me._logPass(me.urlData, {
                eventType: "pc-unreceive-sms-show",
              });
              me._ownerDialog && me._ownerDialog.hide("unHide");
              me.getElement("unRecevieSmsSendCode").innerHTML = rsp.captcha;
              me.getElement("upsms-dialog-code") && (me.getElement("upsms-dialog-code").innerText = rsp.captcha);
              var smsSendNumber = "";
              if (rsp.channelCode) {
                smsSendNumber = rsp.channelCode + "";
                smsSendNumber = smsSendNumber.replace(/\s/g, "").replace(/(.{4})/g, "$1 ");
              }
              me.getElement("unRecevieSmsSendNumber").innerHTML = smsSendNumber;
              me.getElement("upsms-dialog-smsNumber") &&
                (me.getElement("upsms-dialog-smsNumber").innerText = smsSendNumber);
              me.$show(me.getElement("unRecevieSmsTip"));
              preUpSms && me.$hide(me.getElement("unRecevieSmsTip"));
              passport[me.$getId("spareWData")].sign = rsp.channelId;
              me.config.unReceiveSmsCode = rsp.captcha || "";
              me._createChannel(rsp.channelId);
              // 标记channel来源，区分轮询
              me.config.channelType = "unReceiveSms";
            } else {
              passport[me.$getId("spareWData")].unicast && clearTimeout(passport[me.$getId("spareWData")].unicast);
              me._setSmsGeneralError(rsp.errmsg);
              me.getElement("upsms-dialog-wrap") && baidu(me.getElement("upsms-dialog-wrap")).hide();
            }
          },
          error: function () {},
        });
      },

      _setChannel: function () {
        var me = this;
        // 标记channel来源，区分轮询
        me.config.channelType = "qrcode";
        var logPage = "traceId:" + me.urlData.traceId + ",logPage:" + me.urlData.page;

        var data = {
          apiver: "v3",
          tt: new Date().getTime(),
          tpl: me.config.product || "",
          logPage: logPage,
        };
        var container = me.getElement("qrcodeContent");
        // passport.spareWData = passport.spareWData || {};
        passport[me.$getId("spareWData")] = passport[me.$getId("spareWData")] || {};
        baidu(".Qrcode-status-con", container).hide();
        me.$show(me.getElement("QrcodeMain"));
        me.$show(me.getElement("QrcodeTitle"));
        me.config.isOpenWxMiniQrcodeLogin && me.$show(me.getElement("changeWxQrLoginTitle"));
        me.$show(me.getElement("QrcodeDownload"));

        passport[me.$getId("spareWData")].loadQrcode = setTimeout(function () {
          if (me.config.qrcode_style) {
            baidu(".Qrcode-status-con", container).hide();
            baidu(".refresh-title", container).hide();
            baidu(".refresh-loadout", container).show();
            me.$hide(me.getElement("QrcodeTitle"));
            me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
            me.$hide(me.getElement("QrcodeDownload"));
            me.$show(me.getElement("QrcodeRefresh"));
          }
        }, 5 * 1000);
        var url =
          (passport.requestDomain || me._domain.https) +
          "/v2/api/getqrcode?lp=pc&qrloginfrom=" +
          me.config.qrloginfrom +
          "&gid=" +
          (me.guideRandom || "");
        // thirdCheckAgreement=1且showAgreeCheckbox=1时传给server
        if (+me.config.thirdCheckAgreement === 1 && me.config.showAgreeCheckbox) {
          url += "&checkprotocal=1";
        }
        // qrcodePrivacyLink有值时拼入二维码请求
        if (me.config.qrcodePrivacyLink) {
          url += "&qrcodePrivacyLink=" + encodeURIComponent(me.config.qrcodePrivacyLink);
        }
        // 增加网盘的小程序参数
        if (me.config.moreParams && me.config.moreParams.qrext_clientid) {
          url += "&qrext_clientid=" + me.config.moreParams.qrext_clientid;
        }
        baidu.ajax({
          url: url,
          dataType: "jsonp",
          data: data,
          success: function (args) {
            clearTimeout(passport[me.$getId("spareWData")].loadQrcode);
            clearTimeout(passport[me.$getId("spareWData")].unicast);
            if (+args.errno === 50000) {
              baidu(".Qrcode-status-con").hide();
              me.$show(me.getElement("QrcodeRisk"));
              return;
            }
            passport[me.$getId("spareWData")].channelimg =
              (window.location ? window.location.protocol : document.location.protocol) + "//" + args.imgurl;
            passport[me.$getId("spareWData")].sign = args.sign;
            var img = baidu(".tang-pass-qrcode-img", container);
            for (var i = 0, len = img.length; i < len; i++) {
              img.get(i).src = passport[me.$getId("spareWData")].channelimg;
            }
            setTimeout(function () {
              me._createChannel(passport[me.$getId("spareWData")].sign);
            }, 300);
            passport[me.$getId("spareWData")].timer = setTimeout(
              function () {
                clearTimeout(passport[me.$getId("spareWData")].unicast);
                me._stopChannel();
                baidu(".Qrcode-status-con", container).hide();
                baidu(".refresh-title", container).hide();
                baidu(".refresh-timeout", container).show();
                me.$hide(me.getElement("QrcodeTitle"));
                me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
                me.$hide(me.getElement("QrcodeDownload"));
                me.$show(me.getElement("QrcodeRefresh"));
                return;
              },
              10 * 59 * 1000,
            );
          },
          error: function () {
            clearTimeout(passport[me.$getId("spareWData")].loadQrcode);
            clearTimeout(passport[me.$getId("spareWData")].unicast);
            if (me.config.qrcode_style) {
              baidu(".Qrcode-status-con", container).hide();
              me.$hide(me.getElement("QrcodeTitle"));
              me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
              me.$hide(me.getElement("QrcodeDownload"));
              me.$show(me.getElement("QrcodeError"));
            }
          },
        });
      },

      _stopChannel: function () {
        var me = this;
        passport[me.$getId("spareWData")] = passport[me.$getId("spareWData")] || {};
        passport[me.$getId("spareWData")].sign = "";
        clearTimeout(passport[me.$getId("spareWData")].timer);
        clearTimeout(passport[me.$getId("spareWData")].loadQrcode);
        clearTimeout(passport[me.$getId("spareWData")].unicast);
      },

      _createChannel: function (sign) {
        var me = this;
        var qrcodeSign = sign;
        var container = me.getElement("qrcodeContent");
        var qrcodeInit = baidu(".tang-pass-qrcode-init", container).get(0);
        var qrcodeImg = baidu(".tang-pass-qrcode-img", qrcodeInit).get(0);
        var data = {
          apiver: "v3",
          tt: new Date().getTime(),
        };
        passport[me.$getId("spareWData")] = passport[me.$getId("spareWData")] || {};
        // 上行短信轮询不展示二维码错误提示
        if (!me.config.unReceiveSmsCode) {
          passport[me.$getId("spareWData")].unicast = setTimeout(function () {
            baidu(".Qrcode-status-con", container).hide();
            me.$hide(me.getElement("QrcodeTitle"));
            me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
            me.$hide(me.getElement("QrcodeDownload"));
            me.$show(me.getElement("QrcodeError"));
          }, 35 * 1000);
        }
        if (!passport[me.$getId("spareWData")].sign) {
          return false;
        }
        baidu.ajax({
          url:
            (passport.requestDomain || me._domain.https) +
            "/channel/unicast?channel_id=" +
            passport[me.$getId("spareWData")].sign +
            "&gid=" +
            (me.guideRandom || "") +
            "&tpl=" +
            me.config.product +
            "&_sdkFrom=1",
          dataType: "jsonp",
          data: data,
          success: function (d) {
            if (!me.config.unReceiveSmsCode) {
              clearTimeout(passport[me.$getId("spareWData")].unicast);
            }
            if (d.channel_v) {
              try {
                d.channel_v = eval("(" + d.channel_v + ")");
              } catch (e) {
                d.channel_v = {};
              }
            } else {
              d.channel_v = {};
            }
            if (d.channel_v.u) {
              d.channel_v.u = decodeURIComponent(d.channel_v.u);
            }
            if (me.config.channelType === "unReceiveSms" && +d.errno === 0 && +d["channel_v"] === 1) {
              me._logPass(me.urlData, {
                eventType: "pc-smslogin-unreceive-send-" + (me.config.isPwd && me.config.isPwd < 0 ? "reg" : "login"),
              });
              document.getElementById(me.$getId("smsVerifyCode")).value = me.config.unReceiveSmsCode;
              me.config.unReceiveSmsCode = "";
              me._submitSmsForm(sign);
              me.$hide(me.getElement("unRecevieSmsTip"));
              me._ownerDialog && me._ownerDialog.show();
            } else if (+d.errno === 0 && +d.channel_v.status === 0) {
              clearInterval(passport[me.$getId("spareWData")].timer);
              var data = {
                bduss: d.channel_v.v,
                u: encodeURIComponent(d.channel_v.u || me.config.u),
                loginVersion: "v5",
                qrcode: 1,
                tpl: me.config.product ? me.config.product : "",
                maskId: me.config.maskId || "",
                fileId: me.config.fileId || "",
              };
              // passport.data.jsonp('/v2/api/bdusslogin?bduss=' + d.channel_v.v + '&u=' + encodeURIComponent(d.channel_v.u || me.config.u) + '&loginVersion=v4&qrcode=1&tpl=' + (me.config.product ? me.config.product : '')).success(function (args) {
              passport.data
                .jsonp("/v3/login/main/qrbdusslogin?v=" + new Date().getTime(), data)
                .success(function (args) {
                  args.loginType = "qrcode";
                  if (args.errInfo.no == 0) {
                    // 记录登录方式
                    me.storageSupport(function () {
                      localStorage.setItem("passLoginType", "qrcode");
                    });
                    me.config.loginType = "qrLogin";

                    var returnValue = me.fireEvent("loginSuccess", {
                      rsp: Object.assign({}, { type: "qrLogin" }, args),
                    });
                    if (!returnValue) {
                      return;
                    }
                    // yy扫码登录回调
                    if (window.YY_TPL_CONFIG.indexOf(me.config.product + ",") > -1) {
                      passport.data.setContext(baidu.extend({}, me.config));
                      passport.data
                        .getAccessToken({
                          client: "",
                          clientfrom: "",
                          tpl: me.config.product,
                          // 'client_id': me.config.clientId
                          client_id: "Bxr73Efet8HjR5Tr0HqcgDKr",
                        })
                        .success(function (accessReps) {
                          var data = accessReps.data;
                          var returnValue = me.fireEvent("yyLoginSuccess", {
                            rsp: data,
                          });
                          if (!returnValue) {
                            return;
                          }
                          window.location.href =
                            rsp.channel_v.u + "?accesstoken=" + data.access_token + "&expire=" + data.expires_in;
                        });
                    }
                    window.location.href = args.data.u;
                  } else {
                    me.fireEvent("loginError", {
                      rsp: Object.assign({}, { type: "qrLogin" }, args),
                    });
                    me._logPass(me.urlData, {
                      eventType: "pc-qrLogin-submit-error",
                      errno: args.errInfo.no,
                    });
                  }
                });
            } else {
              if (d.errno == 0 && d.channel_v.status == "1") {
                // 扫码成功
                if (me.config.qrcode_style) {
                  baidu(".Qrcode-status-con", container).hide();
                  me.$hide(me.getElement("QrcodeTitle"));
                  me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
                  me.$hide(me.getElement("QrcodeDownload"));
                  me.$show(me.getElement("QrcodeSuccess"));
                  me._callBehaviorCallback({
                    type: "qr-code-suc",
                  });
                }
                if (passport[me.$getId("spareWData")].sign && qrcodeSign == passport[me.$getId("spareWData")].sign) {
                  me._createChannel(qrcodeSign);
                }
              } else if (d.errno == 0 && d.channel_v.status == "2") {
                clearInterval(passport[me.$getId("spareWData")].timer);
                var returnValue = me.fireEvent("loginQrcodeCancel");
                if (!returnValue) {
                  return;
                }
                if (me.config.qrcode_style) {
                  baidu(".Qrcode-status-con", container).hide();
                  baidu(".refresh-title", container).hide();
                  baidu(".refresh-timeout", container).show();
                  me.$hide(me.getElement("QrcodeTitle"));
                  me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
                  me.$hide(me.getElement("QrcodeDownload"));
                  me.$show(me.getElement("QrcodeRefresh"));
                } else {
                  if (qrcodeImg) {
                    qrcodeImg.src = passport.apiDomain.staticDomain + "/passApi/img/loading.gif";
                  }
                  me._setChannel();
                  // passport[me.$getId('spareWData')].timer = setInterval(function () {
                  //     me._setChannel();
                  // }, 10 * 60 * 1000);
                }
              } else if (
                passport[me.$getId("spareWData")].sign &&
                qrcodeSign == passport[me.$getId("spareWData")].sign
              ) {
                me._createChannel(qrcodeSign);
              }
            }
          },
          error: function () {
            if (!me.config.unReceiveSmsCode) {
              clearTimeout(passport[me.$getId("spareWData")].unicast);
            }
            if (me.config.qrcode_style) {
              baidu(".Qrcode-status-con", container).hide();
              me.$hide(me.getElement("QrcodeTitle"));
              me.config.isOpenWxMiniQrcodeLogin && me.$hide(me.getElement("changeWxQrLoginTitle"));
              me.$hide(me.getElement("QrcodeDownload"));
              me.$show(me.getElement("QrcodeError"));
            }
          },
        });
      },

      refreshQrcode: function () {
        // 点击刷新二维码
        var me = this;
        clearTimeout(passport[me.$getId("spareWData")].unicast);
        var container = me.getElement("qrcodeContent");
        var qrcodeInit = baidu(".tang-pass-qrcode-init", container).get(0);
        var qrcodeImg = baidu(".tang-pass-qrcode-img", qrcodeInit).get(0);

        if (qrcodeImg) {
          qrcodeImg.src = passport.apiDomain.staticDomain + "/passApi/img/loading.gif";
        }
        me._setChannel();
        // passport[me.$getId('spareWData')].timer = setInterval(function () {
        //     me._setChannel();
        // }, 10 * 60 * 1000);
      },

      _actionQrcode: function () {
        var me = this;
        var container = me.qrcodeDialogDom ? me.qrcodeDialogDom : this.getElement();
        var parent = me._getWDom.parent(container);
        var parents = me._getWDom.parent(parent);

        passport[me.$getId("spareWData")] = passport[me.$getId("spareWData")] || {};
        // config.channelType: 来自短信登录且打开过收不到验证码重新刷新验证码生成新的channelId进行轮询
        if (
          !passport[me.$getId("spareWData")].channelimg ||
          (me.config.channelType && me.config.channelType === "unReceiveSms")
        ) {
          me._setChannel();
          // passport[me.$getId('spareWData')].timer = setInterval(function () {
          //     me._setChannel();
          // }, 10 * 60 * 1000);
        }
      },

      _showQrcode: function () {
        var me = this;
        var container = this.getElement();
        var parent = me._getWDom.parent(container);
        var parents = me._getWDom.parent(parent);
        var btn = baidu(".pass-qrcode-btn", parents).get(0);

        if (me._getWDom.prev(container)) {
          me.$hide(me._getWDom.prev(container));
        }
        me.$hide("choiceuser_article").$hide(container).$show(baidu(".tang-pass-qrcode", parent).get(0));
        if (me._getWDom.next(btn)) {
          me.$hide(btn);
        } else {
          me.$hide(me._getWDom.parent(me._getWDom.parent(btn)));
        }
      },

      /*
    二维码动画显示
    */
      qrcodeAnimationShow: function () {
        var me = this;
        if (me.supportCss3Anim()) {
          baidu(me.getElement("QrcodeMain")).css("margin-left", "0");
        } else {
          baidu(me.getElement("QrcodeMain")).css("margin-left", "0");
        }
        baidu(me.getElement("QrcodeAnimation")).addClass("Qrcode-status-guideAnim");
      },

      /*
    二维码动画隐藏
    */
      qrcodeAnimationHide: function () {
        var me = this;
        baidu(me.getElement("QrcodeAnimation")).removeClass("Qrcode-status-guideAnim");
        if (me.supportCss3Anim()) {
          baidu(me.getElement("QrcodeMain")).removeClass("Qrcode-animation").addClass("Qrcode-animationRight");
        } else {
          baidu(me.getElement("QrcodeMain")).css("margin-left", "99px");
        }
      },
      /*
      检测是否支持css3动画
    */
      supportCss3Anim: function () {
        var test = document.getElementsByTagName("body")[0].style;
        if (typeof test.animation !== "undefined" || typeof test.WebkitAnimation !== "undefined") {
          return true;
        } else {
          return false;
        }
      },
      // 二维码相关事件
      setqrcodeEvent: function () {
        var me = this;
        // 未开启微信小程序登录
        if (!me.config.isOpenWxMiniQrcodeLogin) {
          me.qrcodeAnimationShow();
        }
        baidu(me.getElement("QrcodeErrorfreshBtn")).on("click", function () {
          me.refreshQrcode();
        });
        baidu(me.getElement("QrcodeRefreshBtn")).on("click", function () {
          me.refreshQrcode();
        });
        // 切换为微信小程序扫码
        me.getElement("changeWxQrAppLogin") &&
          baidu(me.getElement("changeWxQrAppLogin")).on("click", function () {
            // 停止百度扫码轮询
            me._stopChannel();
            passport[me.$getId("spareWData")] = {};
            // 一定存在微信扫码容器
            if (me.getElement("wxQrcode")) {
              me.$hide(me.getElement("qrcode"));
              me._setWxMiniQrChannel();
              me.$show(me.getElement("wxQrcode"));
            }
          });
      },
      // 收不到验证码
      getTemplateUnreceiveSms: function () {
        var me = this;
        me.config.unReceiveSmsTempStatus = "none";
        var templateStr =
          '<div id="' +
          me.$getId("unRecevieSmsTip") +
          '" class="pass-unreceiveSms-tip-wrapper" style="display:' +
          me.config.unReceiveSmsTempStatus +
          '">';
        templateStr +=
          '<div class="pass-unreceiveSms-mask"></div>' +
          '<div class="pass-unreceiveSms-tip-wrapper-content"><div class="unreceiveSms-header">' +
          '<h3 class="unreceiveSms-header-h3">' +
          me.lang.cantReceiveSms +
          '</h3><a class="unreceiveSms-header-a" id="' +
          me.$getId("closeUnRecevieSmsTip") +
          '"></a></div>';
        templateStr +=
          '<div class="unreceiveSms-article">' +
          '<div class="pass-unreceiveSms-content">' +
          '<p class="pass-unreceiveSms-p">' +
          me.lang.checkIfBlocked +
          "</p>" +
          '<p class="pass-unreceiveSms-p">' +
          me.lang.overTimeArrive +
          "</p>" +
          '<p class="pass-unreceiveSms-t25" id="' +
          me.$getId("unRecevieSmsReged") +
          '" style="display:none;">' +
          me.lang.reguest +
          (me.config.smsAutoLogin.encryption
            ? ""
            : '<a class="pass-unreceiveSms-btn-tip" id="' +
              me.$getId("unRecevieSmsChangeLogin") +
              '">' +
              me.lang.switchLoginMethod +
              "</a>" +
              me.lang.or +
              "") +
          me.lang.usePhoneSendSms +
          '“<span class="pass-unreceiveSms-btn-tip" id="' +
          me.$getId("unRecevieSmsSendCode") +
          '"></span>”，' +
          me.lang.sendTo +
          "" +
          '<span id="' +
          me.$getId("unRecevieSmsSendNumber") +
          '"></span>' +
          me.lang.shouldLogSuccess +
          '<span class="pass-unreceiveSms-text-tip">' +
          me.lang.introductSms +
          "</span></p>" +
          '<p class="pass-unreceiveSms-t25" id="' +
          me.$getId("unRecevieSmsUnReg") +
          '" style="display:none;">' +
          me.lang.rule1 +
          me.lang.rule2 +
          '<span class="pass-unreceiveSms-text-tip">' +
          me.lang.introductSms +
          "</span></p>" +
          "</div></div></div>";
        return templateStr;
      },
      _getTemplateSms: function () {
        var me = this;
        var isShow = "none";
        var templateStr =
          '<div id="' +
          me.$getId("sms") +
          '" class="tang-pass-login tang-pass-sms ' +
          (me.config.smsAutoLogin.encryption ? "tang-pass-sms-encryption" : "") +
          '" style="display:' +
          isShow +
          '">';
        var hiddenFields = {
          u: me.config.u,
          staticPage: me.config.staticPage,
          tpl: me.config.product ? me.config.product : "",
          idc: me.config.idc ? me.config.idc : "",
          isdpass: "1",
          gid: me.guideRandom || "",
          switchuname: "",
          smsCodeString: "",
          smsVcodesign: "",
          smsVcodestr: "",
          subpro: me.config.subpro,
          is_voice_sms: me.config.is_voice_sms,
          voice_sms_flag: 0,
          mkey: me.config.mkey || "",
        };
        templateStr += '<p class="tang-pass-sms-tip">' + (me.config.smsText || me.lang.verifyEqualLogin) + "</p>";
        templateStr += '<form id="' + me.$getId("smsForm") + '" class="pass-form-smsForm" method="POST">';
        templateStr += me._getHiddenField(hiddenFields, "smsHiddenFields");
        templateStr +=
          '<div id="' +
          me.$getId("smsPhoneWrapper") +
          '" class="pass-form-item pass-form-item-smsPhone' +
          (!me.foreignMobile ? "" : " pass-form-item-PhoneCountry") +
          '">' +
          (me.foreignMobile
            ? '<label for="' +
              me.$getId("smsPhone") +
              '" id="' +
              me.$getId("smsPhoneCountryLabel") +
              '" class="pass-label pass-label-smsPhone' +
              '" data-countryCode="' +
              (me.isEnVersion ? "001" : "") +
              '">' +
              (me.isEnVersion ? "+1" : "+86") +
              "</label>"
            : "") +
          '<input id="' +
          me.$getId("smsPhone") +
          '" type="text" name="username" class="pass-text-input pass-text-input-smsPhone"' +
          (me.config.smsAutoLogin.encryption ? "disabled" : "") +
          "/>" +
          '<span id="' +
          me.$getId("smsPhoneTip") +
          '" class="pass-item-tip pass-item-tip-smsPhone" style="display:none"><span id="' +
          me.$getId("smsPhoneTipText") +
          '"></span></span>' +
          (me.foreignMobile ? '<ul id="' + me.$getId("smsCountryList") + '" class="pass-country-list"></ul>' : "") +
          "</div>";
        templateStr +=
          '<p id="' +
          me.$getId("smsVerifyCodeWrapper") +
          '" class="pass-form-item pass-form-item-smsVerifyCode">' +
          '<input id="' +
          me.$getId("smsVerifyCode") +
          '" type="text" name="password" class="pass-text-input pass-text-input-smsVerifyCode" maxlength="6" />' +
          '<button id="' +
          me.$getId("smsTimer") +
          '" class="pass-item-timer">' +
          me.lang.smsVerifyCode +
          "</button>" +
          '<span id="' +
          me.$getId("smsVerifyCodeTip") +
          '" class="pass-item-tip pass-item-tip-smsVerifyCode" style="display:none"><span id="' +
          me.$getId("smsVerifyCodeTipText") +
          '"></span></span>' +
          "</p>";
        templateStr +=
          '<p id="' +
          me.$getId("smsUnReceiveWrapper") +
          '" class="pass-form-item pass-form-item-unrecevie" style="display:none;">' +
          '<span id="' +
          me.$getId("smsUnReceiveTipWrapper") +
          '" class="pass-sms-unreceive-tips">' +
          me.lang.unReceiveSmsCode +
          "</span></p>";
        templateStr +=
          '<p id="' +
          me.$getId("smsErrorWrapper") +
          '" class="pass-generalErrorWrapper">' +
          '<span id="' +
          me.$getId("smsError") +
          '" class="pass-generalError"></span>' +
          "</p>";
        templateStr +=
          '<p id="' +
          me.$getId("smsMemberPassWrapper") +
          '" style="display:none;visibility: hidden;">' +
          '<input id="' +
          me.$getId("smsMemberPass") +
          '" type="checkbox" name="memberPass"' +
          (me.config.memberPass ? ' checked="checked"' : "") +
          " />" +
          "</p>";
        templateStr +=
          '<p id="' +
          me.$getId("unbindingSMSTipsWrapper2") +
          '" class="pass-unbindingSMSTipsWrapper pass-unbindingSMSTipsWrapper2" style="display: none;">' +
          me.lang.unbindingSmsTip +
          '<span id="' +
          me.$getId("unbindingSMSTipsBtn2") +
          '" class="pass-unbindingSMSTipsBtn">' +
          me.lang.unblocking +
          "</span></p>";
        templateStr +=
          '<p id="' +
          me.$getId("certErrorWrapper2") +
          '" class="pass-unbindingSMSTipsWrapper pass-unbindingSMSTipsWrapper2" style="display: none;">' +
          me.lang.unbindingSmsTip1 +
          '<span id="' +
          me.$getId("certErrorBtn2") +
          '" class="pass-unbindingSMSTipsBtn">' +
          me.lang.unblocking +
          "</span></p>";
        if (me.config.is_voice_sms != 1) {
          templateStr +=
            '<p id="' +
            me.$getId("smsSubmitWrapper") +
            '" class="pass-form-item pass-form-item-submit">' +
            '<input id="' +
            me.$getId("smsSubmit") +
            '" type="submit" value="' +
            me.lang.login +
            '" class="pass-button pass-button-submit' +
            (me.config.showAgreeCheckbox ? ' pass-button-full-disabled" disabled />' : '"/>') +
            '<span id="' +
            me.$getId("agreementSmsWrapper") +
            '" class="tang-pass-sms-agreement pass-link">' +
            (me.config.showAgreeCheckbox
              ? '<input name="smsIsAgree" id="' +
                me.$getId("smsIsAgree") +
                '" type="checkbox" class="pass-checkbox-input pass-checkbox-isAgree" autocomplete="off" />' +
                '<label for="' +
                me.$getId("smsIsAgree") +
                '" style="padding-left: 5px;">' +
                me.lang.agree +
                "</label>"
              : me.lang.agree) +
            '<a target="_blank" href="' +
            me.constant.PROTOCAL_URL +
            '" style="display: inline-block;margin-bottom: 6px">' +
            (me.config.protocalText || me.lang.baiduUserProtocal) +
            "</a> " +
            (me.customizeAnd ? "、" : me.lang.and) +
            '<a target="_blank" href="' +
            me.constant.PERSONAL_PROTOCAL_URL +
            '">' +
            (me.config.userpersonalText || me.lang.baiduUserPersonal) +
            "</a>" +
            (me.config.showChildAgree
              ? '、<a target="_blank" href="' +
                me.constant.CHILD_AGREEMENT_URL +
                '">' +
                me.lang.childAgreement +
                "</a> "
              : "") +
            "</span>" +
            // + (me.config.smsAutoLogin.encryption ? '' : ('<a id="' + me.$getId('sms_btn_back')
            //     + '" class="pass-sms-link pass-sms-link-back pass-link">' + me.lang.backToLogin + '</a>'))
            "</p>";
        } else {
          templateStr +=
            '<p id="' +
            me.$getId("smsSubmitWrapper") +
            '" class="pass-form-item pass-form-item-submit">' +
            '<input id="' +
            me.$getId("smsSubmit") +
            '" type="submit" value="' +
            me.lang.login +
            '" class="pass-button pass-button-submit' +
            (me.config.showAgreeCheckbox ? ' pass-button-full-disabled" disabled />' : '"/>') +
            '<input  type="button" ' +
            'style="border:none;background:none;margin-top:12px;cursor:pointer;color:#2e7fdb;font-size:12px"' +
            ' class="pass-is_voice"  id="getVoiceSms" value="' +
            me.lang.voiceCode +
            '" />' +
            // + (me.config.smsAutoLogin.encryption ? '' : ('<a id="' + me.$getId('sms_btn_back')
            //     + '" class="pass-sms-link pass-sms-link-back pass-link">' + me.lang.backToLogin + '</a>'))
            '</br><span id="' +
            me.$getId("agreementSmsWrapper") +
            '" class="tang-pass-sms-agreement pass-link">' +
            (me.config.showAgreeCheckbox
              ? '<input name="smsIsAgree" id="' +
                me.$getId("smsIsAgree") +
                '" type="checkbox" class="pass-checkbox-input pass-checkbox-isAgree" autocomplete="off" />' +
                '<label for="' +
                me.$getId("smsIsAgree") +
                '" style="padding-left: 5px;">' +
                me.lang.agree +
                "</label>"
              : me.lang.agree) +
            '<a target="_blank" href="' +
            me.constant.PROTOCAL_URL +
            '" style="display: inline-block;margin-bottom: 6px">' +
            (me.config.protocalText || me.lang.baiduUserProtocal) +
            "</a> " +
            (me.customizeAnd ? "、" : me.lang.and) +
            '<a target="_blank" href="' +
            me.constant.PERSONAL_PROTOCAL_URL +
            '">' +
            (me.config.userpersonalText || me.lang.baiduUserPersonal) +
            "</a>" +
            (me.config.showChildAgree
              ? '、<a target="_blank" href="' +
                me.constant.CHILD_AGREEMENT_URL +
                '">' +
                me.lang.childAgreement +
                "</a> "
              : "") +
            "</span>" +
            "</p>";
        }
        templateStr += "</form>";
        templateStr += "</div>";
        return templateStr;
      },

      _setEventSms: function () {
        var me = this;
        var container = this.getElement();
        var parent = me._getWDom.parent(container);
        var parents = me._getWDom.parent(parent);
        var smsLogin = baidu("#" + me.$getId("sms"), parents).get(0);
        var baiduInput = baidu(".pass-text-input", smsLogin);

        if (me.foreignMobile) {
          baidu(me.getElement("smsPhoneCountryLabel")).on("click", function (evt) {
            var _smsCountryList = me.getElement("smsCountryList");
            if (_smsCountryList && _smsCountryList.style.display != "block") {
              me.$show(_smsCountryList);
              baidu(me.getElement("smsPhoneCountryLabel")).addClass("pass-label-code-up");
            } else if (_smsCountryList) {
              me.$hide(_smsCountryList);
              baidu(me.getElement("smsPhoneCountryLabel")).removeClass("pass-label-code-up");
            }
            me._selectCountryList(me.getElement("smsPhoneWrapper"));
            evt.preventDefault();
          });
        }

        baiduInput.on("mouseover", function (e) {
          var returnValue = me.fireEvent("fieldMouseover", {
            ele: baidu(this),
          });
          if (!returnValue) {
            return;
          }
          baidu(this).addClass(me.constant.HOVER_CLASS);
        });
        baiduInput.on("mouseout", function (e) {
          var returnValue = me.fireEvent("fieldMouseout", {
            ele: baidu(this),
          });
          if (!returnValue) {
            return;
          }
          baidu(this).removeClass(me.constant.HOVER_CLASS);
        });
        baiduInput.on("keydown", function (e) {
          if (+e.keyCode === 13) {
            if (e && e.preventDefault) {
              e.preventDefault();
            }
            var smsIsAgree = baidu("#" + me.$getId("smsIsAgree"));
            var agreementChecked = smsIsAgree && smsIsAgree.get(0) && smsIsAgree.get(0).checked;
            // 判断是否有勾选协议
            if (me.config.showAgreeCheckbox && !agreementChecked) {
              return;
            } else {
              // 发送点击短信登录按钮统计
              var autoStatisticObj = {
                eventType: "pc-smslogin-submit-click",
              };
              me._logPass(me.urlData, autoStatisticObj);

              me._submitSmsForm();
            }
          }
        });

        var baiduSmsPhone = baidu(me.getElement("smsPhone"));
        baiduSmsPhone.on("focus", function (e) {
          if (!me.initialized) {
            me._initApi();
          }
          var returnValue = me.fireEvent("fieldFocus", {
            ele: baidu(this),
          });
          if (!returnValue) {
            return;
          }
          baidu(this).addClass(me.constant.FOCUS_CLASS);
          baidu(this).removeClass(me.constant.ERROR_CLASS);
          var smsRegPromptWrapper = document.getElementById(me.$getId("smsRegPromptWrapper"));
          var smsloginPromptWrapper = document.getElementById(me.$getId("smsloginPromptWrapper"));
          var smsSecondPromptWrapper = document.getElementById(me.$getId("smsSecondPromptWrapper"));

          if (smsRegPromptWrapper) {
            me.$hide(smsRegPromptWrapper);
          }
          if (smsloginPromptWrapper) {
            me.$hide(smsloginPromptWrapper);
          }
          if (smsSecondPromptWrapper) {
            me.$hide(smsSecondPromptWrapper);
          }
        });
        // 发送短信登录手机号输入统计
        baiduSmsPhone.on("keydown", function () {
          if (!me.isSendSmsInput) {
            me.isSendSmsInput = true;
            var autoStatisticObj = {
              eventType: "pc-smslogin-input",
            };
            me._logPass(me.urlData, autoStatisticObj);
          }
        });

        var baiduCodeInput = baidu(me.getElement("smsVerifyCode"));
        baiduCodeInput.on("focus", function (e) {
          if (!me.initialized) {
            me._initApi();
          }
          var returnValue = me.fireEvent("fieldFocus", {
            ele: baidu(this),
          });
          if (!returnValue) {
            return;
          }
          baidu(this).addClass(me.constant.FOCUS_CLASS);
          baidu(this).removeClass(me.constant.ERROR_CLASS);
        });
        // 发送短信登录短信验证码输入统计
        baiduCodeInput.on("keydown", function () {
          if (!me.isSendSmsCodeInput) {
            me.isSendSmsCodeInput = true;
            var autoStatisticObj = {
              eventType: "pc-smslogin-code-input",
            };
            me._logPass(me.urlData, autoStatisticObj);
          }
        });

        baiduInput.on("blur", function (e) {
          if (this.value) {
            var returnValue = me.fireEvent("fieldBlur", {
              ele: baidu(this),
            });
            if (!returnValue) {
              return;
            }
            if (!me.config.smsAutoLogin.encryption) {
              this.name === "username" ? me._validatorPhoneFn(this) : me._validatorSmsFn(this);
            }
          }
          baidu(this).removeClass(me.constant.FOCUS_CLASS);
        });
        baidu("#" + me.$getId("smsTimer"), smsLogin).on("click", function (evt) {
          me._callBehaviorCallback({
            type: "sms-verify-send",
          });
          // 发送点击发送验证码统计
          var autoStatisticObj = {
            eventType: "pc-smslogin-send-click",
          };
          me._logPass(me.urlData, autoStatisticObj);

          me.config.voice_sms_flag = 0;
          evt.preventDefault();
          me._checkRegPhone();
        });
        baidu("#" + me.$getId("smsSubmit"), smsLogin).on("click", function (evt) {
          // 发送点击短信登录按钮统计
          me._logPass(me.urlData, {
            eventType: "pc-smsLogin-submit",
          });

          if (evt && evt.preventDefault) {
            evt.preventDefault();
          }
          me._submitSmsForm();
        });

        // 勾选协议框
        baidu("#" + me.$getId("smsIsAgree"), smsLogin).on("click", function (evt) {
          var agreeCheck = evt.target.checked;
          var smsSumbit = baidu("#" + me.$getId("smsSubmit"), smsLogin);
          me._callBehaviorCallback({
            checked: agreeCheck,
            type: "sms-check",
          });
          if (agreeCheck) {
            baidu.dom(smsSumbit).removeClass("pass-button-full-disabled").attr("disabled", false);
            me._setSmsGeneralError("");
          } else {
            baidu.dom(smsSumbit).addClass("pass-button-full-disabled").attr("disabled", true);
          }
        });

        baidu(me.getElement("smsSwitchWrapper")).on("click", function () {
          if (/msie 6/i.test(navigator.userAgent) || /msie 7/i.test(navigator.userAgent)) {
            setTimeout(function () {
              me._changeLoginType("sms");
            }, 0);
          } else {
            me._changeLoginType("sms");
          }
          me._logPass(me.urlData, {
            eventType: "user-smsLogin-click",
          });
        });
        baidu(me.getElement("sms_btn_back")).on("click", function () {
          me._changeLoginType("normal");
          if (me.getElement("password")) {
            me._doFocus("password");
          }
        });
        baidu(document.getElementById("getVoiceSms")).on("click", function (evt) {
          me.config.voice_sms_flag = 1;
          evt.preventDefault();
          me._checkRegPhone();
        });
        // 短信故障预案-收不到短信验证码弹窗
        baidu(me.getElement("smsUnReceiveTipWrapper")).on("click", function (evt) {
          me._logPass(me.urlData, {
            eventType: "pc-smslogin-unreceive-sms-btn-click",
          });
          evt.preventDefault();
          if (me.config.isPwd && +me.config.isPwd === 1) {
            me.getChannelForUnReceiveSms();
          } else {
            me._logPass(me.urlData, {
              eventType: "pc-unreceive-sms-show-reg",
            });
            me._ownerDialog && me._ownerDialog.hide("unHide");
            me.$show(me.getElement("unRecevieSmsTip"));
          }
        });
        baidu(me.getElement("closeUnRecevieSmsTip")).on("click", function (evt) {
          evt.preventDefault();
          me.$hide(me.getElement("unRecevieSmsTip"));
          me._ownerDialog && me._ownerDialog.show();
        });
        // 短信故障预案-切换其他登录方式
        baidu(me.getElement("unRecevieSmsChangeLogin")).on("click", function (evt) {
          me._logPass(me.urlData, {
            eventType: "pc-smslogin-unreceive-sms-change-login-click",
          });
          evt.preventDefault();
          me.$hide(me.getElement("unRecevieSmsTip"));
          me._ownerDialog && me._ownerDialog.show();
          var type = "qrcode";
          if (me.config.isPwd && +me.config.isPwd === 1) {
            type = "normal";
          }
          me._changeLoginType(type);
        });
      },
      _setSmsGeneralError: function (msg) {
        var smsEle = this.getElement("smsError");
        smsEle && (smsEle.innerHTML = msg);
      },
      _sendVcode: function (obj, mkdData) {
        var me = obj || this;
        var is_voice_sms = me.config.voice_sms_flag;
        var phone = document.getElementById(me.$getId("smsPhone"));
        var countrycode = me.getElement("smsPhoneCountryLabel")
          ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
          : "";
        var disTime = 60;
        var timer;
        var smsLogin = baidu("#" + me.$getId("sms")).get(0);
        if (!me.config.smsAutoLogin.encryption && !me._validatorPhoneFn(phone)) {
          return;
        }
        baidu("#" + me.$getId("smsRegPromptBtn"), smsLogin).off("click");
        baidu("#" + me.$getId("smsRegPromptBtn"), smsLogin).on("click", function (evt) {
          evt.preventDefault();
        });
        baidu("#" + me.$getId("smsTimer"), smsLogin).off("click");
        baidu("#" + me.$getId("smsTimer"), smsLogin).on("click", function (evt) {
          evt.preventDefault();
        });
        baidu("#" + me.$getId("smsTimer"), smsLogin).removeClass("pass-item-timer");
        baidu("#" + me.$getId("smsTimer"), smsLogin).addClass("pass-item-time-timing");
        var cryptdata = hex_md5(phone.value + "Moonshadow");
        cryptdata = cryptdata.replace(/o/, "ow").replace(/d/, "do").replace(/a/, "ad");
        cryptdata = cryptdata.replace(/h/, "ha").replace(/s/, "sh").replace(/n/, "ns").replace(/m/, "mo");
        var data = {
          gid: me.guideRandom || "",
          username: me._SBCtoDBC(phone.value),
          countrycode: countrycode,
          bdstoken: me.bdPsWtoken,
          tpl: me.config.product ? me.config.product : "",
          loginVersion: "v5",
          flag_code: me.config.voice_sms_flag,
          client: me.config.client,
          mkey: me.config.mkey || "",
          moonshad: cryptdata,
        };
        if (mkdData && mkdData.ds && mkdData.tk) {
          data.ds = mkdData.ds || "";
          data.tk = mkdData.tk || "";
        }
        if (mkdData && mkdData.s && mkdData.k) {
          data.s = mkdData.s || "";
          data.k = mkdData.k || "";
        }
        data.supportdv = "1";
        data.dv = document.getElementById("dv_Input")
          ? document.getElementById("dv_Input").value
          : (window.LG_DV_ARG && window.LG_DV_ARG.dvjsInput) || "";
        // 上行短信失败转下行，标记 isOver
        if (me._isUpSmsOver) {
          data.isOver = 1;
        }
        // 如果产品线传了加密手机号，直接发送加密手机号
        if (me.config.smsAutoLogin.encryption) {
          data.username = me._SBCtoDBC(me.config.smsAutoLogin.encryption);
          data.mobileencryption = 1;
        }
        var domainArea = "";
        if (+me.config.voice_sms_flag === 1) {
          domainArea =
            (passport.requestDomain || window.location.protocol.toLowerCase() + "//wappass.baidu.com") +
            "/wp/api/login/sms?is_voice_sms=" +
            me.config.voice_sms_flag;
        } else {
          domainArea = (passport.requestDomain || me._domain.auto) + "/v2/api/senddpass";
        }
        passport.data.jsonp(domainArea, data).success(function (args) {
          if (
            (+me.config.voice_sms_flag === 0 && +args.data.errno !== 0) ||
            (+me.config.voice_sms_flag === 1 && +args.errInfo.no !== 0)
          ) {
            if (
              +args.data.errno === 18 ||
              +args.data.errno === 19 ||
              +args.errInfo.no === 50020 ||
              +args.errInfo.no === 50021
            ) {
              // args.data.isslide = '1';
              if (args.data.isslide === "1") {
                me.loginPassMkd.initVcode();
                me._ownerDialog && me._ownerDialog.hide("unHide");
              } else {
                var confirmVerifyCodeImgSrc = me.constant.VERIFYCODE_URL_PREFIX + args.data.vcodestr;
                me.getElement("smsHiddenFields_smsVcodesign").value = args.data.vcodesign;
                me.getElement("smsHiddenFields_smsVcodestr").value = args.data.vcodestr;
                if (confirmSmsVerifyWidget) {
                  // TODO 切换验证码图片
                  me.getElement("confirmVerifyCodeImg").src = confirmVerifyCodeImgSrc;
                  me.getElement("confirmVerifyCode").value = "";
                  me._ownerDialog && me._ownerDialog.hide("unHide");
                  confirmSmsVerifyWidget.show();
                } else {
                  passport._load(me._domain.auto + "/passApi/js/uni_wrapper.js", true, function () {
                    window.confirmSmsVerifyWidget = passport.pop.init({
                      type: "confirmWidget",
                      tangram: true,
                      titleText: me.lang.safety,
                      width: 490,
                      apiOpt: {
                        Continue: me.lang.confirmText,
                        contentHTML:
                          '<p class="pass-confirm-verifyWidget-msg">' +
                          me.lang.fillPicture +
                          "</p>" +
                          '<p class="pass-confirm-verifyWidget-imgWrapper">' +
                          '<input type="text" class="pass-text-input' +
                          ' pass-confirm-verifyWidget-verifyCode" id="' +
                          me.$getId("confirmVerifyCode") +
                          '" name="confirmVerifyCode" value="" />' +
                          '<img src="' +
                          confirmVerifyCodeImgSrc +
                          '" title="" class="pass-confirm-verifyWidget-verifyCode-img" id="' +
                          me.$getId("confirmVerifyCodeImg") +
                          '" />' +
                          '<a href="#" class="pass-confirm-verifyWidget-imgChange" id="' +
                          me.$getId("confirmVerifyCodeChange") +
                          '">' +
                          me.lang.captchaChange +
                          "</a>" +
                          '<span class="pass-confirm-verifyWidget-error" id="' +
                          me.$getId("confirmVerifyCodeError") +
                          '"></span>' +
                          "</p>",
                      },
                      onRender: function (evt) {
                        baidu(confirmSmsVerifyWidget.getElement("confirmWidget_footer")).addClass(
                          "pass-confirm-verifyWidget-bottom",
                        );
                        me.config.hasPlaceholder &&
                          me._getPlaceholder([
                            {
                              label: "confirmVerifyCode",
                              placeholder: "verifyCode",
                            },
                          ]);
                        baidu(me.getElement("confirmVerifyCodeChange")).on("click", function () {
                          baidu(me.getElement("confirmVerifyCodeImg")).attr(
                            "src",
                            me.constant.VERIFYCODE_URL_PREFIX +
                              me.getElement("smsHiddenFields_smsVcodestr").value +
                              "&v=" +
                              new Date().getTime(),
                          );
                        });
                        baidu(me.getElement("confirmVerifyCode")).on("keyup", function () {
                          baidu(me.getElement("confirmVerifyCode")).removeClass("pass-text-input-error");
                          baidu(me.getElement("confirmVerifyCodeError")).hide();
                          baidu(me.getElement("confirmVerifyCodeError")).get(0).innerHTML = "";
                        });
                        baidu(me.getElement("confirmVerifyCode")).on("change", function () {
                          me.getElement("confirmVerifyCode").value = me
                            .getElement("confirmVerifyCode")
                            .value.replace(/\s/g, "");
                        });
                      },
                      onConfirmClose: function (evt) {
                        baidu(me.getElement("confirmVerifyCodeError")).hide();
                        baidu(me.getElement("confirmVerifyCodeError")).get(0).innerHTML = "";
                        confirmSmsVerifyWidget.hide();
                        me._ownerDialog && me._ownerDialog.show();
                      },
                      onConfirmCancel: function (evt) {},
                      onConfirmContinue: function (evt) {
                        if (me.getElement("confirmVerifyCode").value == "") {
                          baidu(me.getElement("confirmVerifyCode")).addClass("pass-text-input-error");
                          baidu(me.getElement("confirmVerifyCodeError")).show();
                          baidu(me.getElement("confirmVerifyCodeError")).get(0).innerHTML = me.lang.confirmVerCodeEmpty;
                          return;
                        }
                        var dvInput = document.getElementById("dv_Input");
                        var cryptdata = hex_md5(phone.value + "Moonshadow");
                        cryptdata = cryptdata.replace(/o/, "ow").replace(/d/, "do").replace(/a/, "ad");
                        cryptdata = cryptdata
                          .replace(/h/, "ha")
                          .replace(/s/, "sh")
                          .replace(/n/, "ns")
                          .replace(/m/, "mo");
                        var data = {
                          gid: me.guideRandom || "",
                          username: me._SBCtoDBC(phone.value),
                          countrycode: countrycode,
                          bdstoken: me.bdPsWtoken,
                          tpl: me.config.product ? me.config.product : "",
                          vcodestr: me.getElement("smsHiddenFields_smsVcodestr").value,
                          vcodesign: me.getElement("smsHiddenFields_smsVcodesign").value,
                          verifycode: me._SBCtoDBC(me.getElement("confirmVerifyCode").value),
                          flag_code: me.config.voice_sms_flag,
                          loginVersion: "v5",
                          mkey: me.config.mkey || "",
                          moonshad: cryptdata,
                          dv: dvInput ? dvInput.value : (window.LG_DV_ARG && window.LG_DV_ARG.dvjsInput) || "",
                        };
                        // 如果产品线传了加密手机号，直接发送加密手机号
                        if (me.config.smsAutoLogin.encryption) {
                          data.username = me._SBCtoDBC(me.config.smsAutoLogin.encryption);
                          data.mobileencryption = 1;
                        }
                        // 上行短信失败转下行，标记 isOver
                        if (me._isUpSmsOver) {
                          data.isOver = 1;
                        }
                        var domainArea = "";
                        if (me.config.voice_sms_flag == 1) {
                          domainArea =
                            (passport.requestDomain || window.location.protocol.toLowerCase() + "//wappass.baidu.com") +
                            "/wp/api/login/sms?is_voice_sms=" +
                            me.config.voice_sms_flag;
                        } else {
                          domainArea = (passport.requestDomain || me._domain.auto) + "/v2/api/senddpass";
                        }
                        passport.data.jsonp(domainArea, data).success(function (rsp) {
                          var baiduCodeError = baidu(me.getElement("confirmVerifyCodeError"));
                          if (
                            (me.config.voice_sms_flag == 0 && rsp.data.errno == 0) ||
                            (me.config.voice_sms_flag == 1 && rsp.errInfo.no == 0)
                          ) {
                            if (me.config.voice_sms_flag == 1) {
                              var voiceTime = 15;
                              document.getElementById("getVoiceSms").disabled = true;
                              timer = setInterval(function () {
                                var elementVoiceSms = document.getElementById("getVoiceSms");
                                if (!elementVoiceSms) {
                                  return;
                                }
                                var $getId = me.$getId("smsTimer");
                                if (--voiceTime == 0) {
                                  clearInterval(timer);
                                  var baiduSmsTimer = baidu("#" + $getId, smsLogin);
                                  baiduSmsTimer.removeClass("pass-item-time-timing");
                                  baiduSmsTimer.addClass("pass-item-timer");
                                  elementVoiceSms.disabled = false;
                                  elementVoiceSms.value = me.lang.resend + me.lang.vCode;
                                  /* eslint-disable max-len */
                                  document.getElementById($getId).innerHTML = me.lang.resend;
                                  voiceTime = 60;
                                } else {
                                  var baiduSmsTimer = baidu("#" + $getId, smsLogin);
                                  baiduSmsTimer.addClass("pass-item-time-timing");
                                  baiduSmsTimer.removeClass("pass-item-timer");
                                  elementVoiceSms.value = me.lang.Sent + voiceTime + "s";
                                }
                              }, 1000);
                            } else {
                              document.getElementById(me.$getId("smsTimer")).disabled = true;
                              timer = setInterval(function () {
                                var $getId = me.$getId("smsTimer");
                                var elementSmsTimer = document.getElementById($getId);
                                if (!elementSmsTimer) {
                                  return;
                                }
                                if (--disTime == 0) {
                                  clearInterval(timer);
                                  var baiduSmsTimer = baidu("#" + $getId, smsLogin);
                                  baiduSmsTimer.removeClass("pass-item-time-timing");
                                  baiduSmsTimer.addClass("pass-item-timer");
                                  document.getElementById(me.$getId("smsTimer")).disabled = false;
                                  elementSmsTimer.innerHTML = me.lang.resend;
                                  disTime = 60;
                                } else {
                                  var baiduSmsTimer = baidu("#" + $getId, smsLogin);
                                  baiduSmsTimer.addClass("pass-item-time-timing");
                                  baiduSmsTimer.removeClass("pass-item-timer");
                                  if (disTime === 49) {
                                    me._logPass(me.urlData, {
                                      eventType: "pc-smslogin-unreceive-btn-show",
                                    });
                                    me.$show(me.getElement("smsUnReceiveWrapper"));
                                  }
                                  elementSmsTimer.innerHTML = disTime + "s";
                                }
                              }, 1000);
                            }
                            baiduCodeError.hide();
                            baiduCodeError.get(0).innerHTML = "";
                            confirmSmsVerifyWidget.hide();
                            me._ownerDialog && me._ownerDialog.show();
                          } else if (rsp.data.errno == 20 || rsp.data.errno == 21) {
                            baidu(me.getElement("confirmVerifyCode")).addClass("pass-text-input-error");
                            baiduCodeError.show();
                            baiduCodeError.get(0).innerHTML = rsp.data.msg;
                            me.getElement("confirmVerifyCodeImg").src =
                              me.constant.VERIFYCODE_URL_PREFIX + rsp.data.vcodestr;
                            me.getElement("confirmVerifyCode").value = "";
                            me.getElement("smsHiddenFields_smsVcodesign").value = rsp.data.vcodesign;
                            me.getElement("smsHiddenFields_smsVcodestr").value = rsp.data.vcodestr;
                          } else if (args.data.errno == 27) {
                            document.location.href =
                              "https://passport.baidu.com/v2/?reg&overseas=" +
                              me.config.overseas +
                              "&tpl=" +
                              me.config.product +
                              "&u=" +
                              encodeURIComponent(me.config.u);
                          } else {
                            baiduCodeError.hide();
                            confirmSmsVerifyWidget.hide();
                            me._ownerDialog && me._ownerDialog.show();
                            if (me.config.voice_sms_flag == 1) {
                              me._setSmsGeneralError(rsp.errInfo.msg);
                            } else {
                              me._setSmsGeneralError(rsp.data.msg);
                            }
                          }
                        });
                      },
                    });
                    me._ownerDialog && me._ownerDialog.hide("unHide");
                    confirmSmsVerifyWidget.show();
                  });
                }
              }
            } else if (+args.data.errno == 50051) {
              me._ownerDialog && me._ownerDialog.hide("unHide");
              baidu(me.getElement("upsms-dialog-wrap")).show();
              me.getChannelForUnReceiveSms("preUpSms");
            } else if (+args.data.errno == 71100) {
              // 新上行短信SDK流程
              // 根据 getphonestatus 结果判断登录/注册场景
              var upSmsScene = me._upSmsIsLogin ? "login" : "register";
              var upSmsPhone = countrycode ? countrycode + phone.value : phone.value;
              me._ownerDialog && me._ownerDialog.hide("unHide");
              me._initUpSmsSdk(upSmsPhone, upSmsScene);
            } else if (+args.data.errno == 50052) {
              if (me.initMkdFlag) {
                me.loginPassMkdv2.show();
              } else {
                me.loginPassMkdv2.init();
              }
              me._ownerDialog && me._ownerDialog.hide("unHide");
            } else if (args.data.errno == 27) {
              document.location.href =
                "https://passport.baidu.com/v2/?reg&overseas=" +
                me.config.overseas +
                "&tpl=" +
                me.config.product +
                "&u=" +
                encodeURIComponent(me.config.u);
            } else if (+me.config.voice_sms_flag !== 1) {
              me._setSmsGeneralError(args.data.msg);
            } else {
              me._setSmsGeneralError(args.errInfo.msg);
            }
            baidu("#" + me.$getId("smsTimer"), smsLogin).addClass("pass-item-timer");
            baidu("#" + me.$getId("smsTimer"), smsLogin).removeClass("pass-item-time-timing");
            document.getElementById(me.$getId("smsTimer")).innerHTML = me.lang.resend;
          } else {
            if (me.config.voice_sms_flag == 1) {
              var voiceTime = 15;
              document.getElementById("getVoiceSms").disabled = true;
              timer = setInterval(function () {
                if (--voiceTime == 0) {
                  clearInterval(timer);
                  var baiduSmsTimer = baidu("#" + me.$getId("smsTimer"), smsLogin);
                  baiduSmsTimer.removeClass("pass-item-time-timing");
                  baiduSmsTimer.addClass("pass-item-timer");
                  document.getElementById("getVoiceSms").disabled = false;
                  document.getElementById("getVoiceSms").value = me.lang.resend + me.lang.vCode;
                  document.getElementById(me.$getId("smsTimer")).innerHTML = me.lang.resend;
                  voiceTime = 15;
                } else {
                  var baiduSmsTimer = baidu("#" + me.$getId("smsTimer"), smsLogin);
                  baiduSmsTimer.addClass("pass-item-time-timing");
                  baiduSmsTimer.removeClass("pass-item-timer");
                  document.getElementById("getVoiceSms").value = me.lang.Sent + voiceTime + "s";
                }
              }, 1000);
            } else {
              document.getElementById(me.$getId("smsTimer")).disabled = true;
              timer = setInterval(function () {
                var elementById = document.getElementById(me.$getId("smsTimer"));
                if (!elementById) {
                  return;
                }
                if (--disTime == 0) {
                  clearInterval(timer);
                  var baiduSmsTimer = baidu("#" + me.$getId("smsTimer"), smsLogin);
                  baiduSmsTimer.removeClass("pass-item-time-timing");
                  baiduSmsTimer.addClass("pass-item-timer");
                  document.getElementById(me.$getId("smsTimer")).disabled = false;
                  elementById.innerHTML = me.lang.resend;
                  disTime = 60;
                } else {
                  var baiduSmsTimer = baidu("#" + me.$getId("smsTimer"), smsLogin);
                  baiduSmsTimer.addClass("pass-item-time-timing");
                  baiduSmsTimer.removeClass("pass-item-timer");
                  if (disTime === 49) {
                    me._logPass(me.urlData, {
                      eventType: "pc-smslogin-unreceive-btn-show",
                    });
                    me.$show(me.getElement("smsUnReceiveWrapper"));
                  }
                  elementById.innerHTML = disTime + "s";
                }
              }, 1000);
            }
          }
          baidu("#" + me.$getId("smsTimer"), smsLogin).on("click", function (evt) {
            me.config.voice_sms_flag = 0;
            evt.preventDefault();
            me._checkRegPhone();
          });
        });
      },
      _validatorPhoneFn: function (field) {
        field.value = field.value.replace(/\s/g, "");
        var me = this;
        function internetOfThingsRuleFn(ele) {
          if (ele.value.length === 13) {
            return !/^(14[01234])\d*$/.test(ele.value);
          }
          return !/^((14[012345689])|(1740)|(1741[12]))\d*$/.test(ele.value);
        }
        if (field.value == "") {
          me._setSmsGeneralError(me.lang.fillPhonePls);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        } else if (
          me.getElement("smsPhoneCountryLabel") &&
          baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") != ""
        ) {
          if (!new RegExp(/^(\d)*$/).test(me._SBCtoDBC(field.value))) {
            me._setSmsGeneralError(me.lang.foreignMobileError);
            baidu(field).addClass(me.constant.ERROR_CLASS);
            return false;
          }
        } else if (!new RegExp(/^1[3456789]\d{9}$/).test(me._SBCtoDBC(field.value))) {
          me._setSmsGeneralError(me.lang.foreignMobileError);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        } else if (!internetOfThingsRuleFn(field) && +me.config.inTMobile) {
          me._setSmsGeneralError(me.lang.internetOfThingsMobileError);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        }
        me._setSmsGeneralError("");
        baidu(field).removeClass(me.constant.ERROR_CLASS);
        return true;
      },
      _validatorSmsFn: function (field) {
        var me = this;
        if (field.value == "") {
          me._setSmsGeneralError(me.lang.fillVerifyCodePls);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        }
        me._setSmsGeneralError("");
        return true;
      },

      _postSmsData: function (data) {
        var me = this;
        data.countrycode = me.getElement("smsPhoneCountryLabel")
          ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
          : "";
        data.token = me.bdPsWtoken;

        data.loginVersion = "v5";
        if (passport.data.traceID) {
          passport.data.traceID.startFlow && passport.data.traceID.startFlow("login");
        }
        data.dv = document.getElementById("dv_Input")
          ? document.getElementById("dv_Input").value
          : (window.LG_DV_ARG && window.LG_DV_ARG.dvjsInput) || "";
        data.bdint_sync_cookie = me.config.isSyncBaiduInt;
        // 上行短信失败转下行，标记 isOver
        if (me._isUpSmsOver) {
          data.isOver = 1;
        }

        data = baidu.extend(data, me.fuid || {});
        passport.data.login(data).success(function (rsp) {
          rsp.loginType = "sms";
          if (rsp.errInfo.no == 0) {
            me.storageSupport(function () {
              localStorage.setItem("passLoginType", "sms");
            });
            me.config.loginType = "smsLogin";

            var returnValue = me.fireEvent("loginSuccess", {
              rsp: Object.assign({}, { type: "sms" }, rsp),
            });

            if (!returnValue) {
              return;
            }
            window.location.href = rsp.data.u;
          } else {
            me.getElement("smsSubmit").style.color = "#fff";
            var returnValue = me.fireEvent("loginError", {
              rsp: Object.assign({}, { type: "sms" }, rsp),
            });
            me._logPass(me.urlData, {
              eventType: "pc-smsLogin-submit-error",
              errno: rsp.errInfo.no,
            });
            if (!returnValue) {
              return;
            }
            if (+rsp.errInfo.no == 400605) {
              // 保存参数用于解禁接口
              me.config.curOptionDatas = data || {};
              me.config.curOptionDatas.ltoken = rsp.data.ltoken || "";
              me.config.curOptionDatas.lstr = rsp.data.lstr || "";
              me.getElement("unbindingSMSTipsWrapper2") && baidu(me.getElement("unbindingSMSTipsWrapper2")).show();
              me.getElement("submit").value = me.lang.login;
              me.getElement("submit").disabled = false;
              // goUnbinding overUnbinding
              me.config.curOption = +rsp.data.hitWlwFreeTime ? "overUnbinding" : "goUnbinding";
              return;
            }
            if (+rsp.errInfo.no === 402013) {
              // 保存参数用于解禁接口
              me.config.certOptions = {
                ltoken: rsp.data.ltoken || "",
                lstr: rsp.data.lstr || "",
              };
              me.getElement("certErrorWrapper2") && baidu(me.getElement("certErrorWrapper2")).show();
              me.getElement("submit").value = me.lang.login;
              me.getElement("submit").disabled = false;
              return;
            }
            // 如果是i标记用户跳申诉
            if (+rsp.errInfo.no === 500072 && !me.isEnVersion) {
              var targetjumpUrl = rsp.data.appealurl || "https://passport.baidu.com/?getpassappeal";
              document.location.href = targetjumpUrl;
              return;
            }

            if (+rsp.errInfo.no === 500072 && me.isEnVersion) {
              me._setSmsGeneralError("There is a security risk cannot login");
              return;
            }

            if (rsp.errInfo.no == 4) {
              me._setSmsGeneralError(me.lang.captchaErr);
            } else {
              me._setSmsGeneralError(rsp.errInfo.msg);
            }

            if (rsp.errInfo.no == 3 || rsp.errInfo.no == 4) {
              // 如果是短信验证码错误，或者需要输入验证码，清空并聚焦验证码栏
              me._clearInput("smsVerifyCode");
            }
          }

          // 发送登录请求完成统计
          var autoStatisticObj = {
            eventType: "pc-smslogin-post-success",
          };
          var urlData = baidu.extend({}, me.urlData);
          urlData.errno = rsp.errInfo.no;
          me._logPass(urlData, autoStatisticObj);
        });
      },

      smsLoginSubmit: function (d) {
        var me = this;
        me._callBehaviorCallback({
          type: "sms-login-submit",
        });
        var d = d || {};
        var data = baidu.form.json(me.getElement("smsForm"));
        if (d.errInfo && d.errInfo.no == 3) {
          passport.data
            .post("/v2/unite-bind", {
              username: d.data.username || "",
              password: data.password,
              countrycode: me.getElement("smsPhoneCountryLabel")
                ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
                : "",
              sms: 1,
              apiver: "v3",
              loginVersion: "v5",
              token: d.data.token || "",
            })
            .success(function (rsp) {
              me._postSmsData(data);
            });
        } else {
          me._postSmsData(data);
        }
      },

      /*短信故障预案中轮询检测到用户已发送短信产生channelV*/
      _submitSmsForm: function (channelId) {
        var me = this;
        var phone = document.getElementById(me.$getId("smsPhone"));
        var vcode = document.getElementById(me.$getId("smsVerifyCode"));
        if (!me.config.smsAutoLogin.encryption && !me._validatorPhoneFn(phone)) {
          phone.focus();
          return;
        }
        if (!me._validatorSmsFn(vcode)) {
          return;
        }
        var returnValue = me.fireEvent("beforeSubmit");
        if (!returnValue) {
          return;
        }
        me.getElement("smsSubmit").style.color = "#9ebef4";
        var data = baidu.form.json(me.getElement("smsForm"));
        data.password = me._SBCtoDBC(data.password);
        data.username = me._SBCtoDBC(data.username);
        data.FP_UID = me._getCookie("FP_UID") || "";
        data.FP_INFO = window.PP_FP_INFO || "";
        data.client = me.config.client;
        data.isupsms = channelId ? 1 : "";
        data.channelid = channelId || "";
        if (me.config.smsAutoLogin.encryption) {
          data.username = encodeURIComponent(encodeURIComponent(me.config.smsAutoLogin.encryption));
          data.mobileencryption = 1;
        }
        function smsLoginFn() {
          me._postSmsData(data);
        }
        if (me.config.source === "phoenix") {
          data.isfromphoenix = 1;
        }

        me.loginConnect(
          {
            username: data.username,
            password: data.password,
            countrycode: me.getElement("smsPhoneCountryLabel")
              ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
              : "",
            smsVcode: data.password,
            isdpass: 1,
            sms: 1,
          },
          {
            fail: function (msg) {
              me._setSmsGeneralError(msg);
            },
          },
          smsLoginFn,
        );
      },

      /**
       * 新上行短信SDK流程
       * 通过独立的 passUpSmsSdk 处理上行短信，SDK负责UI和轮询，
       * 登录态同步由 loginv5 自行处理
       */
      _initUpSmsSdk: function (phoneValue, scene) {
        var me = this;
        var startSdk = function () {
          if (typeof window.passUpSmsSdk === "undefined") {
            // SDK 加载失败，降级走原有流程
            me._ownerDialog && me._ownerDialog.show();
            return;
          }
          var countrycode = me.getElement("smsPhoneCountryLabel")
            ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
            : "";

          // 协议数据透传给 upsms SDK，由 SDK 自行渲染「点击即同意」
          // 协议 URL 优先级：调用方 config 配置 > 网盘 tpl 特殊兜底 > passport 默认 URL
          var isNetdisk = me.config.product === "netdisk";
          var protocalUrl =
            me.config.protocalLink ||
            (isNetdisk ? "https://pan.baidu.com/disk/main#/protocol/duty_mobi" : me.constant.PROTOCAL_URL);
          var personalUrl =
            me.config.userpersonalLink ||
            (isNetdisk
              ? "https://pan.baidu.com/disk/base/protocol#/privacy_policy_addition"
              : me.constant.PERSONAL_PROTOCAL_URL);
          var agreement = [
            {
              text: me.config.protocalText || me.lang.userProtocal,
              url: protocalUrl,
            },
            {
              text: me.config.userpersonalText || me.lang.baiduUserPersonal,
              url: personalUrl,
            },
          ];
          window.passUpSmsSdk.passUpSmsInit({
            tpl: me.config.product || "",
            client: me.config.client || "",
            clientfrom: me.config.clientfrom || "",
            lang: me.config.lang || "zh-CN",
            scene: scene || "register",
            phone: phoneValue,
            overseas: countrycode && countrycode !== "0086" ? 1 : 0,
            agreement: agreement,
            onResult: function (result) {
              switch (result.type) {
                case "success":
                  me._ownerDialog && me._ownerDialog.show();
                  var smsVerifyCodeEl = document.getElementById(me.$getId("smsVerifyCode"));
                  smsVerifyCodeEl && (smsVerifyCodeEl.value = result.smsContent);
                  me._submitSmsForm(result.channelId);
                  break;
                case "fallback":
                  me._ownerDialog && me._ownerDialog.show();
                  me._isUpSmsOver = true;
                  me._checkRegPhone();
                  break;
                case "cancel":
                  me._ownerDialog && me._ownerDialog.show();
                  break;
                case "error":
                  window.passUpSmsSdk && window.passUpSmsSdk.destroyInstance();
                  me._ownerDialog && me._ownerDialog.show();
                  me._setSmsGeneralError((result.extra && result.extra.errmsg) || "");
                  break;
              }
            },
            onClose: function () {
              me._ownerDialog && me._ownerDialog.show();
            },
          });
        };

        if (typeof window.passUpSmsSdk !== "undefined") {
          startSdk();
        } else {
          me._insertScriptW(me.constant.UPSMSSDK_URL, function () {
            startSdk();
          });
        }
      },

      _setEventChoiceUser: function () {
        var me = this;
        // 不用担心第二次自动提交会出现验证码，因为第一次验证全成功，第二次将不出现验证码
        var back = function () {
          baidu(me.getElement()).removeClass("tang-pass-login-hide");
          me.$show(me.getElement()).$hide("choiceuser_article");
        };
        var sub = function (e) {
          baidu(me.getElement()).removeClass("tang-pass-login-hide");
          me.$show(me.getElement()).$hide("choiceuser_article");
          me.submit();
        };
        baidu(me.getElement("choiceuser_btn_username")).on("click", function (e) {
          me.getElement("loginMerge").value = "false";
          sub(e);
        });
        baidu(me.getElement("choiceuser_btn_mobile")).on("click", function (e) {
          me.getElement("isPhone").value = "true";
          me.getElement("loginMerge").value = "false";
          sub(e);
        });
        baidu(me.getElement("choiceuser_btn_back")).on("click", function (evt) {
          evt.preventDefault();
          back();
        });
      },

      _getToken: function (fn) {
        var me = this;
        passport[me.$getId("spareWData")] = passport[me.$getId("spareWData")] || {};
        passport.data
          .getApiInfo({
            apiType: "login",
            loginVersion: "v5",
            gid: me.guideRandom || "",
          })
          .success(function (rsp) {
            me.bdPsWtoken = rsp.data.token;
            fn && fn(me);
          });
      },

      _getRSA: function (callback) {
        var me = this;
        passport.data
          .getRsaKey({
            gid: me.guideRandom || "",
            loginVersion: "v5",
          })
          .success(function (cert) {
            if (!cert.errInfo.no && cert.errInfo.no != 0) {
              cert = cert.data;
            }
            var RSA = new passport.lib.RSA();
            RSA.setKey(cert.pubkey);
            callback &&
              callback({
                RSA: RSA,
                rsakey: cert.key,
              });
          });
      },

      _changeLoginType: function (type, isthirdSence) {
        var me = this;
        var _isthirdSence = isthirdSence || false;
        var container = this.getElement();
        var normalHeight = 0;
        var parent = me._getWDom.parent(container);
        var parents = me._getWDom.parent(parent);
        var qrcode = me.getElement("qrcode");
        var sms = me.getElement("sms");
        var phoenixDom = baidu(".tang-pass-login-phoenix", parents).get(0);
        me._changeTabButtonActivStatus(type);
        // 切换tab清除上次报错信息
        me.clearGeneralError();
        var loginTypeList = !me.config.composeTemplate
          ? {
              normal: {
                $btn: baidu(".pass-normal-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                $ele: baidu(me.getElement("form")).parent(),
              },
              sms: {
                $btn: baidu(".pass-sms-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                $ele: baidu(me.getElement("sms")),
              },
              qrcode: {
                $btn: baidu(".pass-qrcode-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                $ele: baidu(me.getElement("qrcode")),
              },
            }
          : {
              normal: {
                $btn: baidu(".pass-normal-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                $ele: baidu(me.getElement("form")),
              },
              sms: {
                $btn: baidu(".pass-sms-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                $ele: baidu(me.getElement("sms")),
              },
            };
        choiceuser = me.getElement("choiceuser_article");
        type = type || "normal";
        var dotObj = {
          normal: "pwdLogin",
          qrcode: "qrLogin",
          sms: "smsLogin",
        };

        me._logPass(me.urlData, {
          eventType: "pc-" + (dotObj[type] || type) + "-show",
        });
        me._logPass(me.urlData, {
          eventType: "pc-" + type + "login-show",
        });

        if (type === "sms") {
          me.isSendSmsCodeInput = false;
          me.isSendSmsInput = false;
        }
        // 操作一键登录dom
        function changeClientDom() {
          if (me.config.clientShareIsLogin) {
            // 存在账密dom，则需要隐藏
            if (me.getElement("form")) {
              me.$hide(baidu(me.getElement("form"))[0]);
              // 非复合面板，存在客户端互通，优先展示一键互通
              !me.config.composeTemplate && me.$show(baidu(me.getElement("clientonekey"))[0]);
            }
            // 重置报错信息
            me._setClientShareError("");
            loginTypeList.normal = !me.config.composeTemplate
              ? {
                  $btn: baidu(".pass-normal-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                  $ele: baidu(me.getElement("clientonekey")).parent(),
                }
              : {
                  $btn: baidu(".pass-normal-btn", me.getPhoenixElement("pass_phoenix_list_login")),
                  $ele: baidu(me.getElement("clientonekey")),
                };
            me.config.clientShareNoFirst = true;
          }
        }
        // 如果是开启客户端互通pc 一键登录，则账号登录优先展示一键登录
        if (me.config.isOpenClientShare) {
          // 如果开启互通没登录，则去检查客户端登录态且非第一次展示登录框
          if (type === "normal" && !me.config.clientShareIsLogin && me.config.clientShareNoFirst) {
            me._initClientShare(function () {
              changeClientDom();
              changeLoginCallBack();
            });
          } else {
            changeClientDom();
            changeLoginCallBack();
          }
          if (type === "normal" && me.config.clientShareIsLogin) {
            me.config.onRenderClientOnekey &&
              typeof me.config.onRenderClientOnekey === "function" &&
              me.config.onRenderClientOnekey();
          }
        } else {
          changeLoginCallBack();
        }
        /*if (me.config.diaPassLogin && (!me.currentLoginType || me.currentLoginType === 'normal')) {
            if (me.config.userPwdLogin == 1 && parent) {
                normalHeight = Math.max(parent.offsetHeight, 350);
                if (qrcode) {
                    qrcode.style.height = normalHeight + 'px';
                }
            } else if (qrcode) {
                normalHeight = Math.max(qrcode.offsetHeight, 350);
                if (parent) {
                    parent.style.height = normalHeight + 'px';
                }
            }
            if (sms) {
                sms.style.height = normalHeight + 'px';
            }
        }*/
        function changeLoginCallBack() {
          if (choiceuser) {
            me.$hide("choiceuser_article");
          }
          var returnValue = me.fireEvent("changeLoginType", {
            loginType: type,
            currentLoginType: me.currentLoginType || "",
          });
          if (!returnValue) {
            return;
          }
          for (var item in loginTypeList) {
            if (loginTypeList[item].$ele && loginTypeList[item].$ele.length > 0 && !_isthirdSence) {
              if (item == type) {
                me.$show(loginTypeList[item].$ele[0]);
              } else {
                me.$hide(loginTypeList[item].$ele[0]);
              }
            }
          }
          if (type === "qrcode") {
            me._actionQrcode();
          } else if (type === "sms") {
            me.$hide(me.getElement("smsUnReceiveWrapper"));
          }
          me.currentLoginType = type;
        }
      },
      _doFocus: function (ele) {
        var me = this;
        if (!me.config.autoFocus) {
          return;
        }
        if ((typeof ele).toLowerCase() == "string" && me.getElement(ele)) {
          me.getElement(ele).focus();
        } else {
          ele.focus();
        }
      },
      _clearInput: function (type) {
        var me = this;
        var ele = me.getElement(type);
        var eleHolder = me.getElement(type + "_placeholder");
        var eleClear = me.getElement(type + "_clearbtn");
        if (type === "password") {
          // 切换样式
          // 切换input类型
        }
        if (ele) {
          if (eleHolder) {
            me.$show(eleHolder);
          }
          if (eleClear) {
            me.$hide(eleHolder);
          }
          ele.value = "";
          me._doFocus(ele);
        }
      },
      _insertAfterW: function (newNode, oldNode) {
        var me = this;
        var parent = me._getWDom.parent(oldNode);
        if (parent.lastChild == oldNode) {
          parent.appendChild(newNode);
        } else {
          parent.insertBefore(newNode, me._getWDom.next(oldNode));
        }
      },
      _insertBefore: function (newNode, oldNode) {
        var me = this;
        var parent = me._getWDom.parent(oldNode);
        parent.insertBefore(newNode, oldNode);
      },
      initMkd: function () {
        // 新人机
        var me = this;
        me._insertScriptW(me.constant.TOUCHAPIMKD_URL, function () {
          if (window.PassMachine && window.PassMachine.mkd) {
            var PassMkd = window.PassMachine.mkd;
            me.loginPassMkd = new PassMkd({
              maskModule: true,
              // 产品线接入ak
              ak: "1e3f2dd1c81f2075171a547893391274",
              verifySuccessFn: function (data) {
                me.loginPassMkd.removeMask();
                me._ownerDialog && me._ownerDialog.show();
                me.mkdVerifySuccessCallback(data);
              },
              closeFn: function () {
                me._ownerDialog && me._ownerDialog.show();
              },
            });
          }
        });
      },
      initMkdv2: function (callback) {
        // 新人机
        var me = this;
        if (window.sdkMachine) {
          me.loginPassMkdv2 = new window.sdkMachine({
            // 产品线接入ak
            ak: "1e3f2dd1c81f2075171a547893391274",
            deviceType: "wap",
            hasClose: true,
            scene: "login",
            zIndex: 60002, // uc登录组件的层级过高，为避免遮盖，故设置
            errorFn: function () {},
            closeEvent: function () {
              me.loginPassMkdv2.hide();
              me.initMkdFlag = true;
              me._ownerDialog && me._ownerDialog.show();
              return true;
            },
            initApiSuccessFn: function () {
              callback && callback();
            },
            verifySuccessFn: function (data) {
              // me.mkdData = data;
              var mkdDatav2 = {
                s: data.ds,
                k: data.tk,
              };
              me._ownerDialog && me._ownerDialog.show();
              me.mkdVerifySuccessCallback(mkdDatav2);
              me.loginPassMkdv2.hide();
              me.initMkdFlag = true;
            },
          });
        }
      },

      mkdVerifySuccessCallback: function (data) {
        var me = this;
        if (me.currentPath === "mkdDataLoginFn") {
          // pc账密登录出验证码流程
          me.getElement("submit").disabled = true;
          me.getElement("submit").value = me.lang.logining;
          me.getElement("submit").style.color = "#9ebef4";

          me.loginFn(data, me.histroyData);
        } else if (me.currentPath === "mkdDataSendVcode") {
          // pc快速注册出验证码流程
          me._sendVcode(null, data);
        }
      },
      _setPloganondegCookie() {
        var me = this;
        if (me.config.loganondeg === true) {
          me._setCookie("ploganondeg", "1", {
            domain: ".baidu.com",
            secure: true,
            expires: 8 * 365 * 24 * 60 * 60 * 1000, // 有效期八年
          });
        } else {
          // 清除 cookie：写过期时间为过去
          me._setCookie("ploganondeg", "", {
            domain: ".baidu.com",
            secure: true,
            expires: new Date(0), // 1970-01-01
          });
        }
      },

      _insertNoCaptchaScript: function () {
        var me = this;
        me._insertScriptW(me.constant.NOCAPTCHA_URL, function () {});
      },
      _checkCapsLock: function () {
        var me = this;
        var password = baidu(me.getElement("password"));
        password.on("keypress", function (event) {
          var e = event || window.event;
          var keyCode = e.keyCode || e.which;
          var isShift = e.shiftKey || keyCode == 16 || false;
          var caps = document.getElementById(me.$getId("caps"));
          if ((keyCode >= 65 && keyCode <= 90 && !isShift) || (keyCode >= 97 && keyCode <= 122 && isShift)) {
            if (!caps) {
              var span = document.createElement("span");
              span.id = me.$getId("caps");
              // bca-disable-next-line
              span.innerHTML = me.lang.lock;
              var passw = document.getElementById(me.$getId("passwordWrapper"));
              if (passw.style.position == "static") {
                passw.style.position = "relative";
              }
              if (passw.style.zIndex) {
                passw.style.zIndex = passw.style.zIndex + 1;
              } else {
                passw.style.zIndex = 20;
              }
              span.style.cssText =
                "position:absolute;left:60px;clear:both;" +
                "top:25px;width:103px;height:37px;font-size:12px;line-height:45px;" +
                'z-index:20;text-align:center;background:url("' +
                passport.apiDomain.staticDomain +
                '/passApi/img/caps.gif") no-repeat 0 0;';
              passw.appendChild(span);
            } else {
              me.$show(caps);
            }
          } else if (caps) {
            me.$hide(caps);
          }
        });
        password.on("blur", function (event) {
          var caps = document.getElementById(me.$getId("caps"));
          if (caps) {
            me.$hide(caps);
          }
        });
      },
      _checkRegPhone: function (field, callback) {
        var me = this;
        var countrycode = me.getElement("smsPhoneCountryLabel")
          ? baidu(me.getElement("smsPhoneCountryLabel")).attr("data-countrycode") || ""
          : "";
        var phone = document.getElementById(me.$getId("smsPhone"));

        var encryptionMobile = me.config.smsAutoLogin.encryption || "";
        if (!encryptionMobile && !me._validatorPhoneFn(phone)) {
          return;
        }
        passport.data
          .getphonestatus({
            gid: me.guideRandom || "",
            phone: me._SBCtoDBC(encryptionMobile ? encryptionMobile : phone.value),
            loginVersion: "v5",
            countrycode: countrycode,
            mobileencryption: encryptionMobile ? 1 : "",
          })
          .success(function (rsp) {
            var returnValue = me.fireEvent("_checkRegPhone", {
              rsp: rsp,
            });
            if (!returnValue) {
              return;
            }
            if (rsp.errInfo.no == 0 || (+rsp.errInfo.no === 7 && rsp.data.jumpReg === "0")) {
              // 短信故障预案——标记手机号关联的百度账号是否有密码
              me.config.isPwd = +rsp.errInfo.isPwd || 0;
              // 已注册，标记为登录场景
              me._upSmsIsLogin = true;
              baidu(me.getElement("unRecevieSmsUnReg")).hide();
              baidu(me.getElement("unRecevieSmsReged")).show();
              // sendVcodeBefore 第三方调用短信登录，输入手机号检测是否已经绑定对应第三方账号，在此处增加回调，直接登录调用_sendVcode方法，更换手机号则清空输入的手机号。
              if (me.config.sendVcodeBefore && typeof me.config.sendVcodeBefore == "function") {
                me.config.sendVcodeBefore(me, me.mkdDataSendVcode, function () {
                  phone.value = "";
                  phone.focus();
                  return;
                });
              } else if (me.bdPsWtoken) {
                me.mkdDataSendVcode();
              } else {
                me._getToken(me.mkdDataSendVcode);
              }
            } /*else if (+rsp.errInfo.no === 7) {
                baidu(me.getElement('unRecevieSmsReged')).hide();
                baidu(me.getElement('unRecevieSmsUnReg')).show();
                var smsloginPromptWrapper = document.getElementById(me.$getId('smsloginPromptWrapper'));
                var smsPhone = document.getElementById(me.$getId('smsPhoneWrapper'));
                if (!smsloginPromptWrapper) {
                    var div = document.createElement('div');
                    div.id = me.$getId('smsloginPromptWrapper');
                    div.setAttribute('class', 'pass-form-sms-checkphone');
                    div.style.cssText = 'position:absolute;clear:both;color:#826f33;'
                        + 'z-index:999;font-size:12px;width:211px;height:71px;'
                        + 'padding:16px 16px 11px 13px;background:url("'
                        + me._domain.staticFile
                        + '/passApi/img/smsRegphone.png") 0px 0px no-repeat;right:0px;';
                    div.innerHTML = '<p style="margin:0px;padding:0px;line-height:2em;">该手机号未开启手机号登录功能，请使用用户名即密码登录。</p><button id="' + me.$getId('smsLoginPromptBtn') + '" style="background:#2e82ff;border:none;color:#fff;cursor:pointer;height:25px;line-height:25px;width:100px;text-align:center;position:absolute;right:16px;top:66px;" hidefocus=true>用户名密码登录</button>';
                    smsPhone.appendChild(div);
                    me.getElement('smsLoginPromptBtn').focus();
                } else {
                    me.$show(smsloginPromptWrapper);
                    me.getElement('smsLoginPromptBtn').focus();
                }
                baidu(me.getElement('smsLoginPromptBtn')).on('click', function (evt) {
                    if (me.getElement('smsloginPromptWrapper')) {
                        baidu(me.getElement('smsloginPromptWrapper')).hide();
                    }
                    evt.preventDefault();
                    me._changeLoginType('normal');
                    if (me.getElement('password')) {
                        me._doFocus('password');
                    }
                });
            }*/ else if (
              (+rsp.errInfo.no === 3 && rsp.data.jumpReg === "1") ||
              (+rsp.errInfo.no === 7 && rsp.data.jumpReg === "1")
            ) {
              // 未注册，标记为注册场景
              me._upSmsIsLogin = false;
              baidu(me.getElement("unRecevieSmsReged")).hide();
              baidu(me.getElement("unRecevieSmsUnReg")).show();
              var smsSecondPromptWrapper = document.getElementById(me.$getId("smsSecondPromptWrapper"));
              var smsPhone = document.getElementById(me.$getId("smsPhoneWrapper"));
              if (!smsSecondPromptWrapper) {
                var div = document.createElement("div");
                div.id = me.$getId("smsSecondPromptWrapper");
                div.setAttribute("class", "pass-form-sms-checkphone");
                div.style.cssText =
                  "position:absolute;clear:both;color:#826f33;" +
                  "z-index:999;font-size:12px;width:211px;height:71px;box-sizing:content-box;" +
                  'padding:16px 16px 11px 13px;background:url("' +
                  passport.apiDomain.staticDomain +
                  '/passApi/img/smsRegphone.png") 0px 0px no-repeat;right:0px;';
                div.innerHTML =
                  '<p style="margin:0px;padding:0px;line-height:2em;">' +
                  me.lang.registTip +
                  '</p><button id="' +
                  me.$getId("smsSecondPromptBtn") +
                  '" ' +
                  'style="background:#2e82ff;border:none;color:#fff;cursor:pointer;' +
                  "height:25px;line-height:25px;width:100px;text-align:center;position:absolute;" +
                  'right:16px;top:66px;" hidefocus=true>' +
                  me.lang.regist +
                  "</button>";
                smsPhone.appendChild(div);
                me.getElement("smsSecondPromptBtn").focus();
              } else {
                me.$show(smsSecondPromptWrapper);
                me.getElement("smsSecondPromptBtn").focus();
              }
              baidu(me.getElement("smsSecondPromptBtn")).on("click", function (evt) {
                if (me.getElement("smsSecondPromptBtn")) {
                  baidu(me.getElement("smsSecondPromptBtn")).hide();
                }
                evt.preventDefault();
                window.location.href = me.constant.REG_URL;
              });
            } else if (+rsp.errInfo.no === 3) {
              // 短信故障预案——标记手机号未注册
              me.config.isPwd = -1;
              // 未注册，标记为注册场景
              me._upSmsIsLogin = false;
              baidu(me.getElement("unRecevieSmsReged")).hide();
              baidu(me.getElement("unRecevieSmsUnReg")).show();
              var voice = me.config.voice_sms_flag;
              var smsRegPromptWrapper = document.getElementById(me.$getId("smsRegPromptWrapper"));
              var smsPhone = document.getElementById(me.$getId("smsPhoneWrapper"));
              if (!smsRegPromptWrapper) {
                var div = document.createElement("div");
                div.id = me.$getId("smsRegPromptWrapper");
                div.setAttribute("class", "pass-form-sms-checkphone");
                div.style.cssText =
                  "position:absolute;clear:both;color:#826f33;z-index:999;box-sizing:content-box;" +
                  'font-size:12px;width:211px;height:75px;padding:16px 16px 11px 13px;background:url("' +
                  passport.apiDomain.staticDomain +
                  '/passApi/img/smsRegphone.png") 0px 0px no-repeat;right:0px;';
                div.innerHTML =
                  '<p style="margin:0px;padding:0px;line-height:2em;">' +
                  me.lang.registerTip +
                  "</p>" +
                  '<button id="' +
                  me.$getId("smsRegPromptBtn") +
                  '" style="background:#2e82ff;border:none;color:#fff;' +
                  "cursor:pointer;height:25px;line-height:25px;width:68px;text-align:center;" +
                  'position:absolute;right:16px;top:66px;" hidefocus=true>' +
                  me.lang.register +
                  "</button>";
                smsPhone.appendChild(div);
                me.getElement("smsRegPromptBtn").focus();
              } else {
                me.$show(smsRegPromptWrapper);
                me.getElement("smsRegPromptBtn").focus();
              }
              baidu(me.getElement("smsRegPromptBtn")).on("click", function (evt) {
                me.config.voice_sms_flag = voice;
                if (me.getElement("smsRegPromptWrapper")) {
                  baidu(me.getElement("smsRegPromptWrapper")).hide();
                }
                evt.preventDefault();
                // 由于国外账号如果所有手机号状态都是2，国外账号都有区号 那么一定都会跳转注册页面，无法出现多账号列表，所以隐藏掉
                // if (countrycode != '') {
                //     window.location.href = me.constant.REG_URL;
                // } else {
                // sendVcodeBefore 第三方调用短信登录，输入手机号检测是否已经绑定对应第三方账号，在此处增加回调，直接登录调用_sendVcode方法，更换手机号则清空输入的手机号。
                if (me.config.sendVcodeBefore && typeof me.config.sendVcodeBefore == "function") {
                  me.config.sendVcodeBefore(me, me.mkdDataSendVcode, function () {
                    phone.value = "";
                    phone.focus();
                    return;
                  });
                } else if (me.bdPsWtoken) {
                  me.mkdDataSendVcode();
                } else {
                  me._getToken(me.mkdDataSendVcode);
                }
                // }
              });
              me.config.voice_sms_flag = 0;
            } else {
              me.config.voice_sms_flag = 0;
              // 其他错误场景，默认为登录
              me._upSmsIsLogin = true;
              me._setSmsGeneralError(rsp.errInfo.msg);
            }
          });
      },
      // 接收callback函数
      mkdDataSendVcode: function (that) {
        var me = that || this;
        if (me.loginPassMkd && me.loginPassMkd.getDataAsync) {
          me.currentPath = "mkdDataSendVcode";
          var mkdData = {};
          me.loginPassMkd.getDataAsync(function (rsp) {
            mkdData.ds = rsp.ds || "xxx_loginv4";
            mkdData.tk = rsp.tk || "xxx_loginv4";
            if (me.loginPassMkdv2) {
              me.loginPassMkdv2.getDataAsync().then(function (res) {
                mkdData.s = res.ds || "";
                mkdData.k = res.tk || "";
                me._sendVcode(null, mkdData);
              });
            } else {
              me._sendVcode(null, mkdData);
            }
          });
        } else {
          me._sendVcode();
        }
      },
      changeSuggestView: function (opt) {
        var me = this;
        if (me.suggestionDom && opt.list) {
          if (opt.list == "hide") {
            me.$hide(me.suggestionDom);
          } else if (opt.list == "show") {
            me.$show(me.suggestionDom);
          }
        }
        if (me.selectBtn && opt.btn) {
          if (opt.btn == "close") {
            baidu(me.selectBtn).removeClass("open");
            baidu(me.getElement("userName")).addClass("open");
            me.$show(me.selectBtn);
          } else if (opt.btn == "open") {
            baidu(me.selectBtn).addClass("open");
            baidu(me.getElement("userName")).addClass("open");
            me.$show(me.selectBtn);
          } else if (opt.btn == "hide") {
            me.$hide(me.selectBtn);
            baidu(me.getElement("userName")).removeClass("open");
          } else if (opt.btn + "" === "show") {
            me.$show(me.selectBtn);
            baidu(me.getElement("userName")).addClass("open");
          }
          me.$hide(me.selectBtn);
        }
      },
      _suggestion: function (maxlength) {
        var me = this;
        var recordArr = [];
        var $userNameEle = baidu("#" + me.$getId("userName"), me.getElement());
        var suggestionArr = [
          "qq.com",
          "163.com",
          "126.com",
          "sohu.com",
          "sina.com",
          "gmail.com",
          "21cn.com",
          "hotmail.com",
          "vip.qq.com",
          "yeah.net",
          "139.com",
        ];
        var reg = /^([a-zA-Z0-9_.\-+]+)([@]?[a-zA-Z0-9_\-*]*[.]?[a-zA-Z*]*[.]?[a-zA-Z*]*)$/;

        var _appendSuggestTemp = function (text, opt) {
          var current = text;
          if (text.substr(0, text.indexOf("@") - 1).length > opt.maxlength) {
            current = text.substr(0, opt.maxlength - 4) + "…" + text.substr(text.indexOf("@"));
          }
          return baidu(
            '<li class="pass-item-suggsetion" data-select="' +
              text +
              '" data-type="' +
              (opt.ifdelete ? "history" : "normal") +
              '">' +
              current +
              (opt.ifdelete ? '<a data-delete="' + text + '" title="' + me.lang.deleteText + '"></a>' : "") +
              "</li>",
          ).get(0);
        };
        var _rendSelectList = function (u, self) {
          if (!me.suggestionDom) {
            me.suggestionDom = document.createElement("ul");
            me.suggestionDom.id = me.$getId("suggestionWrapper");
            baidu(me.getElement("userNameWrapper")).append(me.suggestionDom);
            baidu(me.suggestionDom).addClass("pass-suggestion-list");
            baidu(me.suggestionDom).on("click", function (evt) {
              var $target = baidu(evt.target);
              var deleteIn = $target.attr("data-delete");
              var currentIn = $target.attr("data-select");
              if (deleteIn) {
                // 删除历史记录
                evt.preventDefault();
                me.suggestionDom.data_delete = true;
                baidu($target.parent()).hide();
                baidu.array(recordArr).remove(deleteIn);
                if (recordArr.length < 1) {
                  me.changeSuggestView({
                    list: " hide",
                    btn: "hide",
                  });
                }

                me._doFocus("userName");
                setTimeout(function () {
                  me.suggestionDom.data_delete = false;
                }, 200);
              } else {
                //autosuggest
                if ($target.attr("data-type") == "history") {
                  // 历史记录autosuggest
                  self.value = currentIn;
                } else {
                  // 邮箱后缀autosuggest
                  me.suggestionDom.data_delete = false;
                  self.value = currentIn || self.value;
                }
                if (me.getElement("userName_placeholder")) {
                  me.$hide("userName_placeholder");
                }
                me.changeSuggestView({
                  list: "hide",
                  btn: "close",
                });
                me._doFocus(self);
                setTimeout(function () {
                  me.setGeneralError("");
                  baidu(self).removeClass("pass-text-input-error");
                  me._doFocus("password");
                }, 100);
                var value = me.getElement("userName").value;
                me._loginCheck(value);
              }
            });
          }

          me.suggestionDom.innerHTML = "";
          me.$show(me.suggestionDom);
          me.suggestionDom.appendChild(u);
          baidu(".pass-item-suggsetion", me.suggestionDom).on("mouseover", function () {
            baidu(".pass-item-suggsetion_hover", me.suggestionDom).removeClass("pass-item-suggsetion_hover");
            baidu(this).addClass("pass-item-suggsetion_hover");
          });
          baidu(".pass-item-suggsetion", me.suggestionDom).on("mouseout", function () {
            baidu(this).removeClass("pass-item-suggsetion_hover");
          });
        };
        var rendLoginHistory = (function () {
          recordArr = me.loginrecord.displayname || [];

          if (recordArr.length > 0) {
            var content = document.createDocumentFragment();
            for (var i = 0, len = recordArr.length; i < len; i++) {
              content.appendChild(_appendSuggestTemp(recordArr[i], { maxlength: maxlength, ifdelete: true }));
            }
            _rendSelectList(content, me.getElement("userName"));
            me.selectBtn = baidu('<span class="pass-item-selectbtn pass-item-selectbtn-userName" ></span>').get(0);
            me.getElement("userNameWrapper").appendChild(me.selectBtn);
            baidu(me.selectBtn).on("click", function (evt) {
              setTimeout(function () {
                if (me.suggestionDom.style.display != "none") {
                  me.changeSuggestView({
                    list: "hide",
                    btn: "close",
                  });
                } else {
                  me.changeSuggestView({
                    list: "show",
                    btn: "open",
                  });
                }
              }, 200);
            });
            me.changeSuggestView({ list: "hide", btn: "show" });
          }
        })();

        $userNameEle.on("keyup", function (evt) {
          if (me.disUnameLogin == 1) {
            // me.disUnameLogin
          } else {
            var u = document.createDocumentFragment();
            var result;
            var self = this;
            var k = 0;

            if (recordArr.length > 0) {
              for (var i = 0, len = recordArr.length; i < len; i++) {
                if (recordArr[i].indexOf(this.value) == 0) {
                  u.appendChild(_appendSuggestTemp(recordArr[i], { maxlength: maxlength, ifdelete: true }));
                  ++k;
                }
              }
            }

            if (recordArr.length < 1 || k < 1) {
              result = reg.exec(this.value);
              if (result && result[2]) {
                for (var i = 0, len = suggestionArr.length; i < len; i++) {
                  if (("@" + suggestionArr[i]).indexOf(result[2]) === 0) {
                    var current = result[1];
                    u.appendChild(_appendSuggestTemp(current + "@" + suggestionArr[i], { maxlength: maxlength }));
                    ++k;
                  }
                }
              }
            }

            if (me.suggestionDom && evt.keyCode !== 38 && evt.keyCode !== 40) {
              me.$hide(me.suggestionDom);
            }
            if (recordArr.length > 1) {
              if (k > 0) {
                if (evt.keyCode !== 38 && evt.keyCode !== 40) {
                  _rendSelectList(u, self);
                }
                if ((evt.keyCode == 38 || evt.keyCode == 40) && me.suggestionDom.style.display != "none") {
                  var child = me.suggestionDom.childNodes;
                  var n = child.length;
                  var p = -1;
                  for (var j = 0; j < n; j++) {
                    if (child[j].className.indexOf("pass-item-suggsetion_hover") > -1) {
                      p = j;
                    }
                  }
                  if (evt.keyCode == 38) {
                    var q = p === -1 ? n - 1 : p === 0 ? n - 1 : p - 1;
                  }
                  if (evt.keyCode == 40) {
                    var q = p === -1 ? 0 : p === n - 1 ? 0 : p + 1;
                  }
                  baidu(".pass-item-suggsetion_hover", me.suggestionDom).removeClass("pass-item-suggsetion_hover");
                  baidu(child[q], me.suggestionDom).addClass("pass-item-suggsetion_hover");
                  var currentIn = baidu(child[q]).attr("data-select");
                  if (baidu(child[q]).attr("data-type") == "history") {
                    self.value = currentIn;
                  } else {
                    self.value =
                      self.value.substr(0, self.value.indexOf("@")) + currentIn.substr(currentIn.indexOf("@"));
                  }
                  if (me.getElement("userName_placeholder")) {
                    me.$hide("userName_placeholder");
                  }
                }
              }
            }
          }
        });

        $userNameEle.on("keydown", function (evt) {
          if ((evt.keyCode == 13 || evt.keyCode == 9) && me.suggestionDom && me.suggestionDom.style.display != "none") {
            me.changeSuggestView({
              list: "hide",
              btn: "close",
            });
            me._doFocus("password");
            evt.preventDefault();
            evt.stopPropagation();
          }
        });

        $userNameEle.on("blur", function () {
          setTimeout(function () {
            if (me.suggestionDom && !me.suggestionDom.data_delete) {
              // data_delete是一个稳定suggestion的开关，防止点击clearbtn或者placeholder时suggestdom出现闪动的情况
              me.changeSuggestView({
                list: "hide",
                btn: "close",
              });
            }
          }, 150);
        });

        $userNameEle.on("focus", function (evt) {
          me.changeSuggestView({
            list: "show",
            btn: "open",
          });
        });
      },

      /**
       * @description _initCountryCode 初始化国家区域代码已经国家名称
       * @param {Object} domField 插入海外手机号dom
       */
      _initCountryCode: function (domField) {
        var me = this;
        var countryTempStr = me.isEnVersion
          ? ""
          : '<li class="pass-item-country"><span class="pass-country-code" data-countryCode="">+86</span>' +
            me.lang.mainland +
            "</li>";
        var data = me.countryCodeList || {};
        var countryLen = data.length;
        if (countryLen <= 0) {
          return;
        }
        for (var i = 0; i < countryLen; i++) {
          countryTempStr +=
            '<li class="pass-item-country"><span class="pass-country-code" data-countryCode=' +
            data[i].code +
            ">+" +
            data[i].code.substring(2) +
            "</span>" +
            data[i].name +
            "</li>";
        }
        baidu(domField).html(countryTempStr);
      },

      _initCustomizeAgreement: function () {
        var me = this;
        var customizeStr = "";
        var Length = me.config.customAgreementList.length;
        for (var i = 0; i < Length; i++) {
          var style = i === 1 ? ' style="display: inline-block; margin-bottom: 6px"' : "";
          customizeStr +=
            '、<a target="_blank" href="' +
            me.config.customAgreementList[i].link +
            '"' +
            style +
            ">" +
            me.config.customAgreementList[i].text +
            "</a>";
        }
        var agreementWrapper = baidu(me.getElement("agreementWrapper")) && baidu(me.getElement("agreementWrapper"))[0];
        var agreementSmsWrapper =
          baidu(me.getElement("agreementSmsWrapper")) && baidu(me.getElement("agreementSmsWrapper"))[0];
        function updateContent(wrapper) {
          if (wrapper) {
            wrapper.innerHTML += customizeStr;
          }
        }
        updateContent(agreementWrapper);
        updateContent(agreementSmsWrapper);
        var agreementQrcodeWrapperEle = me.getElement("agreementQrcodeWrapper");
        if (agreementQrcodeWrapperEle) {
          updateContent(baidu(agreementQrcodeWrapperEle)[0]);
        }
      },

      /**
       * @description _getCountryCode 获取国家区域代码已经国家名称
       * @param {fn} callback callback函数
       */
      _getCountryCode: function (callback) {
        var me = this;
        var data = {
          apiver: "v3",
          loginVersion: "v5",
          subpro: me.config.subpro,
          lang: me.config.lang,
          tpl: me.config.product || (passport && passport.tpl),
        };
        passport.data
          .jsonp((passport.requestDomain || me._domain.https) + "/v2/?securitygetcountrycode", data)
          .success(function (rsp) {
            if (rsp.data.country.length > 0) {
              me.countryCodeList = rsp.data.country;
              if (me.getElement("foreignCountryList")) {
                me._initCountryCode(me.getElement("foreignCountryList"));
              }
              if (me.getElement("smsCountryList")) {
                me._initCountryCode(me.getElement("smsCountryList"));
              }
              callback && callback();
            }
          });
      },

      _selectCountryList: function (domField) {
        var me = this;
        var _domField = baidu(domField);
        var _countryList = _domField.find(".pass-country-list").eq(0);
        var _labelCode = _domField.find(".pass-label");
        if (_labelCode.length == 0) {
          return;
        }
        _countryList.on("click", function (evt) {
          var target = evt.target;
          if (target.tagName.toLowerCase() === "li") {
            _labelCode.eq(0).html(baidu(target).find("span.pass-country-code").eq(0).html());
            _labelCode
              .eq(0)
              .attr("data-countryCode", baidu(target).find("span.pass-country-code").eq(0).attr("data-countryCode"));
          } else if (target.tagName.toLowerCase() === "span") {
            _labelCode.eq(0).html(baidu(target).html());
            _labelCode.eq(0).attr("data-countryCode", baidu(target).attr("data-countryCode"));
          }
          me.$hide(_countryList[0]);
          _labelCode.eq(0).removeClass("pass-label-code-up");
          if (
            domField === me.getElement("foreignMobileWrapper") &&
            me.getElement("foreignMobile") &&
            me.getElement("foreignMobile").value
          ) {
            me._validatorforeignmobileFn(me.getElement("foreignMobile"));
          } else if (
            domField === me.getElement("smsPhoneWrapper") &&
            me.getElement("smsPhone") &&
            me.getElement("smsPhone").value
          ) {
            me._validatorPhoneFn(me.getElement("smsPhone"));
          }
          evt.preventDefault();
        });
        _countryList.on("mouseover", function (evt) {
          var target = evt.target;
          if (target.tagName.toLowerCase() === "li") {
            _domField.find(".pass-item-country-hover").removeClass("pass-item-country-hover");
            baidu(target).addClass("pass-item-country-hover");
          } else if (target.tagName.toLowerCase() === "span") {
            _domField.find(".pass-item-country-hover").removeClass("pass-item-country-hover");
            baidu(target).parent("li.pass-item-country").addClass("pass-item-country-hover");
          }
        });
        _countryList.on("mouseout", function (evt) {
          var target = evt.target;
          if (target.tagName.toLowerCase() === "li") {
            baidu(target).removeClass("pass-item-country-hover");
          } else if (target.tagName.toLowerCase() === "span") {
            baidu(target).parent("li.pass-item-country").removeClass("pass-item-country-hover");
          }
        });
        // 点击其他区域收起国家代码list
        baidu("html").on("click", function (evt) {
          var target = evt.target;
          if (!_labelCode) {
            return false;
          }
          if (
            baidu(target).attr("id") != baidu(me.getElement("foreignMobileLabel")).attr("id") &&
            baidu(target).attr("id") != baidu(me.getElement("smsPhoneCountryLabel")).attr("id")
          ) {
            setTimeout(function () {
              me.$hide(_countryList[0]);
              _labelCode.eq(0).removeClass("pass-label-code-up");
            }, 200);
          }
        });
      },

      _setForeignMobileEvent: function () {
        var me = this;
        if (me.getElement("foreignMobileLabel")) {
          baidu(me.getElement("foreignMobileLabel")).on("click", function (evt) {
            var _CountryList = me.getElement("foreignCountryList");
            if (_CountryList && _CountryList.style.display !== "block") {
              me.$show(_CountryList);
              baidu(me.getElement("foreignMobileLabel")).addClass("pass-label-code-up");
            } else if (_CountryList) {
              me.$hide(_CountryList);
              baidu(me.getElement("foreignMobileLabel")).removeClass("pass-label-code-up");
            }
            me._selectCountryList(me.getElement("foreignMobileWrapper"));
            evt.preventDefault();
          });
        }
        if (me.getElement("foreignMobile")) {
          baidu(me.getElement("foreignMobile")).on("blur", function () {
            if (this.value) {
              var returnValue = me.fireEvent("fieldBlur", {
                ele: baidu(this),
              });
              if (!returnValue) {
                return;
              }
              me._validatorforeignmobileFn(this);
            }
            baidu(this).removeClass(me.constant.FOCUS_CLASS);
          });
          baidu(me.getElement("foreignMobile")).on("focus", function () {
            if (!me.initialized) {
              me._initApi();
            }
            var returnValue = me.fireEvent("fieldFocus", {
              ele: baidu(this),
            });
            if (!returnValue) {
              return;
            }
            baidu(this).addClass(me.constant.FOCUS_CLASS);
            baidu(this).removeClass(me.constant.ERROR_CLASS);
          });
        }

        if (me.getElement("foreignMobileLink")) {
          baidu(me.getElement("foreignMobileLink")).on("click", function (evt) {
            me.$hide(me.getElement("userNameWrapper"));
            me.$hide(me.getElement("smsSwitchWrapper"));
            me.$hide(me.getElement("foreignMobileLink"));
            baidu(me.getElement("userName")).removeClass(me.constant.ERROR_CLASS);
            me.setGeneralError("");
            me.getElement("password").value = "";
            me.$show(me.getElement("foreignMobileWrapper"));
            me.$show(me.getElement("foreignMobileMsg"));
            me.$show(me.getElement("foreignMobileBackWrapper"));
            me.internation = true;
            evt.preventDefault();
            var autoStatisticObj = {
              eventType: "foreignMobileLinkClick",
            };
            me._logPass(me.urlData, autoStatisticObj);
          });
        }

        if (me.getElement("foreignMobileBackWrapper")) {
          baidu(me.getElement("foreignMobileBackWrapper")).on("click", function (evt) {
            me.$hide(me.getElement("foreignMobileWrapper"));
            me.$hide(me.getElement("foreignMobileMsg"));
            me.$hide(me.getElement("foreignMobileBackWrapper"));
            baidu(me.getElement("foreignMobile")).removeClass(me.constant.ERROR_CLASS);
            me.setGeneralError("");
            me.getElement("password").value = "";
            me.$show(me.getElement("userNameWrapper"));
            me.$show(me.getElement("smsSwitchWrapper"));
            me.$show(me.getElement("foreignMobileLink"));
            me.internation = false;
            evt.preventDefault();
          });
        }
      },

      _validatorforeignmobileFn: function (field) {
        var me = this;
        if (field.value == "") {
          me.setGeneralError(me.lang.fillPhonePls);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        } else if (
          me.getElement("foreignMobileLabel") &&
          baidu(me.getElement("foreignMobileLabel")).attr("data-countrycode") != ""
        ) {
          if (!new RegExp(/^(\d)*$/).test(me._SBCtoDBC(field.value))) {
            me.setGeneralError(me.lang.foreignMobileError);
            baidu(field).addClass(me.constant.ERROR_CLASS);
            return false;
          }
        } else if (!new RegExp(/^1[3456789]\d{9}$/).test(me._SBCtoDBC(field.value))) {
          me.setGeneralError(me.lang.foreignMobileError);
          baidu(field).addClass(me.constant.ERROR_CLASS);
          return false;
        }
        me.setGeneralError("");
        baidu(field).removeClass(me.constant.ERROR_CLASS);
        return true;
      },

      _rendPhoenixbtn: function () {
        var me = this;
        var $phoenixEle = baidu(me.getPhoenixElement("pass_phoenix_list_login"));
        $phoenixEle.on("click", function (evt) {
          var $target = baidu(evt.target);
          var changeTo = $target.attr("data-type");
          var title = $target.attr("data-title");
          me._callBehaviorCallback({
            name: title,
            type: "phoenix-login",
          });
          if (changeTo) {
            me._changeLoginType(changeTo, true);
          }
        });
      },

      setMakeText: function (str) {
        var me = this;
        var makeDom = me.getElement("MakeTextWrapper");
        var str = str
          .replace(/</g, "&lt;")
          .replace(/>/g, "&gt;")
          .replace(/"/g, "&quot;")
          .replace(/\x22/g, "&quot;")
          .replace(/\x27/g, "&#39;");
        if (!makeDom) {
          return;
        }
        if (str) {
          makeDom.style.display = "";
          makeDom.innerHTML = str;
        } else {
          makeDom.style.display = "none";
          makeDom.innerHTML = "";
        }
      },
      // 账号密码登录事件处理
      _setPasswordAgreeEvent: function () {
        var me = this;

        // 勾选协议框
        baidu(me.getElement("isAgree")).on("click", function (evt) {
          var agreeCheck = evt.target.checked;
          var smsSumbit = me.getElement("submit");
          me._callBehaviorCallback({
            checked: agreeCheck,
            type: "password-check",
          });
          if (agreeCheck) {
            baidu.dom(smsSumbit).removeClass("pass-button-full-disabled").attr("disabled", false);
          } else {
            baidu.dom(smsSumbit).addClass("pass-button-full-disabled").attr("disabled", true);
          }
        });
      },
      /**
       * @description render 渲染组件到页面
       * @function
       * @name magic.passport.login#render
       * @grammar magic.passport.login#render(id)
       * @param {String} id 渲染到的容器的 id
       */
      render: function (id) {
        var me = this;
        function handleRender() {
          if (!me.getElement()) {
            me.$mappingDom("", id || document.body);
          }
          var passLoginType = "";
          var target = baidu(me.getElement()); //baidu('#'+id);
          // me.getElement().parentNode.style.width = "400px"
          target.addClass(me.constant.CONTAINER_CLASS);
          var headerTemplate = me._getHeaderTemplate();
          // 判断是否是客户端互通登录，且有登录态的情况
          if (me.config.isOpenClientShare && me.config.clientShareIsLogin) {
            var template = me._getClientShareTemplate();
          } else {
            var template = me._getTemplate();
          }
          document.body.insertAdjacentHTML("afterbegin", me._getIrregularField("tipToast"));
          // 判断复合面板
          var qrcodeTemplate = me._getTemplateQrcode();
          var wxminiTemplate = me._getTemplateWxMiniQrcode();
          if (me.config.composeTemplate) {
            // var composeParentDom = baidu(me.getElement().parentNode);
            target.addClass("compose-style");
            // 添加两个外层box 横向布局
            var composeTemplate = me._setComposeTemplate();
            var composeLeftNode = baidu(composeTemplate).get(0);
            var composeRightNode = baidu(composeTemplate).get(1);
            // if (+me.config.hideLogo !== 1) {
            composeLeftNode.appendChild(baidu(headerTemplate).get(0));
            // }
            if (me.config.qrcode === 3 || +me.config.qrcodeLogin === 1) {
              if (
                Object.prototype.toString.call(me.config.qrcodeDom).toLowerCase() === "[object function]" &&
                me.config.diaPassLogin &&
                +me.config.qrcode === 3
              ) {
                composeLeftNode.appendChild(baidu(qrcodeTemplate).get(0));
                if (
                  ((!passLoginType || passLoginType === "qrcode" || (passLoginType === "sms" && !me.smsConfig)) &&
                    +me.config.userPwdLogin !== 1) ||
                  me.config.qrcodeLogin === 1
                ) {
                  setTimeout(function () {
                    me._actionQrcode();
                  }, 500);
                }
                me.getElement("qrcode").style.display = "";
              } else {
                composeLeftNode.appendChild(baidu(qrcodeTemplate).get(0));
                // me.getElement('qrcode').style.display = 'block';
                if (
                  ((!passLoginType || passLoginType === "qrcode" || (passLoginType === "sms" && !me.smsConfig)) &&
                    +me.config.userPwdLogin !== 1) ||
                  me.config.qrcodeLogin === 1
                ) {
                  me._setEventQrcode();
                }
                me.$show(baidu(me.getElement("qrcode")));
                baidu(me.getElement("qrcode")).show();
              }
            }
            if (me.config.isOpenWxMiniQrcodeLogin) {
              composeLeftNode.appendChild(baidu(wxminiTemplate).get(0));
              me._setWxMiniQrChannel();
            }
            // 插入头部
            composeRightNode.appendChild(baidu(headerTemplate).get(1));
            // 短信
            var smsTemplate = me._getTemplateSms();
            var unReceiveSmsTemplate = me.getTemplateUnreceiveSms();
            document.getElementsByTagName("body")[0].appendChild(baidu(unReceiveSmsTemplate).get(0));
            // 合并
            composeRightNode.appendChild(baidu(template).get(0));
            composeRightNode.appendChild(baidu(smsTemplate).get(0));
            target.get(0).appendChild(composeLeftNode);
            target.get(0).appendChild(composeRightNode);
            me._initCustomizeAgreement();
            me._setRowDistance();
            me._setPasswordAgreeEvent();
            me._setEventSms();
            me.setqrcodeEvent();
            if (me.config.isOpenWxMiniQrcodeLogin) {
              me._setWxMiniQrcodeEvent();
            }
          }

          if (me.config.makeText) {
            me.setMakeText(me.config.makeText);
          }

          if (me.config.setWebToClient) {
            var webToClintTemplate = me._getIrregularField("webtoclint");
            me._insertAfterW(baidu(webToClintTemplate).get(0), me.getElement());
          }
          if (!me.config.smsAutoLogin.encryption) {
            var footerBarTemplate = me.getTemplateFooterBar();
            me._insertAfterW(baidu(footerBarTemplate).get(0), target.get(0));
          }
          if (me.config.authsiteLogin && me.config.authsiteLogin.length > 0) {
            var otherTemplate = me._getTemplateOther();
            // bca-disable-next-line
            me.getElement("PhoenixItem").innerHTML = otherTemplate;
            me._rendPhoenixbtn();
            me._authSiteW();
          }
          me.storageSupport(function () {
            passLoginType = localStorage.getItem("passLoginType");
          });
          if (!me.config.composeTemplate) {
            target.get(0).appendChild(baidu(template).get(0));
            // me._setPasswordAgreeEvent();

            // 二维码
            // 默认扫码qrcode === 3 并且 产品线没配置账密 userPwdLogin !== 1
            // 产品线配置的二维码 qrcodeLogin === 1
            if (me.config.qrcode === 3 || +me.config.qrcodeLogin === 1) {
              if (
                Object.prototype.toString.call(me.config.qrcodeDom).toLowerCase() === "[object function]" &&
                me.config.diaPassLogin &&
                +me.config.qrcode === 3
              ) {
                me.qrcodeDialogDom = me.config.qrcodeDom();
                me.qrcodeDialogDom.appendChild(baidu(qrcodeTemplate).get(0));
                if (
                  ((!passLoginType || passLoginType === "qrcode" || (passLoginType === "sms" && !me.smsConfig)) &&
                    +me.config.userPwdLogin !== 1) ||
                  me.config.qrcodeLogin === 1
                ) {
                  setTimeout(function () {
                    me._actionQrcode();
                  }, 500);
                }
                me.getElement("qrcode").style.display = "";
              } else {
                me._insertAfterW(baidu(qrcodeTemplate).get(0), target.get(0));
                if (
                  ((!passLoginType || passLoginType === "qrcode" || (passLoginType === "sms" && !me.smsConfig)) &&
                    +me.config.userPwdLogin !== 1) ||
                  me.config.qrcodeLogin === 1
                ) {
                  me._setEventQrcode();
                }
              }
              me.setqrcodeEvent();
            }

            if (me.smsConfig) {
              // 动态密码
              var smsStrTemplate = me._getTemplateSms();
              var unReceiveSmsTemplates = me.getTemplateUnreceiveSms();
              me._insertAfterW(baidu(smsStrTemplate).get(0), target.get(0));
              document.getElementsByTagName("body")[0].appendChild(baidu(unReceiveSmsTemplates).get(0));
              // me._setEventSms();
            }
            me._initCustomizeAgreement();
            me._setRowDistance();
            me._setPasswordAgreeEvent();
            if (me.smsConfig) {
              me._setEventSms();
            }
          }
          // 传入加密手机号和打码手机号，默认为短信登录
          if (me.config.smsAutoLogin.mobile && me.config.smsAutoLogin.encryption) {
            passLoginType = "sms";
          }

          // 如果没有上次登录方式就不用切换了 因为代码初始化默认就是账密
          if (
            (passLoginType === "normal" || (passLoginType === "sms" && me.smsConfig)) &&
            me.config.userPwdLogin !== 1 &&
            me.config.qrcodeLogin !== 1
          ) {
            me._changeLoginType(passLoginType);
          }
          // 如果上次登录是二维码，但是扫码登录已隐藏，则降级到密码
          if (passLoginType === "qrcode" && (me.config.qrcode === 0 || +me.config.qrcodeLogin === 0)) {
            passLoginType = "normal";
          }
          if (!me.config.composeTemplate) {
            // 隐藏百度logo
            // if (+me.config.hideLogo !== 1) {
            me._insertBefore(baidu(headerTemplate).get(0), target.get(0));
            // }
            // 插入头部
            me._insertBefore(baidu(headerTemplate).get(1), target.get(0));
          }
          var disableJump = (me.config.qrcodeCfg && me.config.qrcodeCfg.disableJump) || false;
          if (disableJump) {
            document.querySelector(".pass-link").addEventListener(
              "click",
              function (event) {
                event.preventDefault();
              },
              false,
            );
          }
          // 如果开启了客户端互通且有登录态情况
          if (me.config.isOpenClientShare && me.config.clientShareIsLogin) {
            passLoginType = "normal";
            me._changeLoginType(passLoginType);
          }
          // 初始化切换方法
          me.setLoginTypeTabBar();
          if (passLoginType !== null) {
            me._changeTabButtonActivStatus(passLoginType);
          } else {
            var defaultLoginType;
            if (me.smsConfig && !me.config.userPwdLogin) {
              defaultLoginType = "sms";
            }
            if (!(me.config.qrcode === 3 || +me.config.qrcodeLogin === 1) && me.config.userPwdLogin === 1) {
              defaultLoginType = "normal";
            }
            if (me.config.qrcode === 3 || +me.config.qrcodeLogin === 1 || me.config.composeTemplate) {
              // 如果有互通且客户端有登录态，则优先互通
              if (me.config.isOpenClientShare && me.config.clientShareIsLogin) {
                defaultLoginType = "normal";
              } else {
                defaultLoginType = "qrcode";
              }
            }
            defaultLoginType !== "qrcode" && me._changeLoginType(defaultLoginType);
            me._changeTabButtonActivStatus(defaultLoginType);
          }
          setTimeout(function () {
            // 猎豹浏览器自动记住表单项有问题，导致loginMerge参数异常，加延时重置
            me.getElement("loginMerge") && (me.getElement("loginMerge").value = "true");
          }, 200);

          if (me.config.hasPlaceholder) {
            var rendList = [
              {
                label: "userName",
                placeholder: me.config && me.config.hideEmailAndPhoneText ? "userNameTitle" : "userName",
              },
              {
                label: "password",
                placeholder: "password",
              },
              {
                label: "verifyCode",
                placeholder: "verifyCode",
              },
            ];
            if (me.smsConfig) {
              if (!me.config.smsAutoLogin.encryption) {
                rendList.push({
                  label: "smsPhone",
                  placeholder: me.config && me.config.diaPassLogin ? "smsPhoneMsg" : "smsPhone",
                });
              }
              rendList.push({
                label: "smsVerifyCode",
                placeholder: "smsVerifyPlaceholder",
              });
            }
            if (me.foreignMobile) {
              rendList.push({
                label: "foreignMobile",
                placeholder: me.config && me.config.diaPassLogin ? "smsPhoneMsg" : "smsPhone",
              });
            }
            me._getPlaceholder(rendList);
          }
          if (me.foreignMobile) {
            me._getCountryCode();
            me._setForeignMobileEvent();
          }
          var autoStatisticObj = {
            eventType: "loginShow",
          };
          me._logPass(me.urlData, autoStatisticObj);
          // 如果开启互通客户端登录
          if (me.config.isOpenClientShare) {
            me._setClientShareEvent();
          }
          /**
           * @description render 渲染组件到页面
           * @event
           * @name magic.passport.login#render
           * @grammar magic.passport.login#render
           */
          var returnValue = me.fireEvent("render");
          if (!returnValue) {
            return;
          }
          var mise = true;
          if (
            navigator.appName === "Microsoft Internet Explorer" &&
            (navigator.appVersion.split(";")[1].replace(/[ ]/g, "") === "MSIE6.0" ||
              navigator.appVersion.split(";")[1].replace(/[ ]/g, "") === "MSIE7.0")
          ) {
            mise = false;
          }

          // 上行短信弹窗 不是点击“未收到短信”出现的弹窗
          me.initPreUpsms();
          // 物联网手机号解禁
          me.initUnbindingSMS();
          // 风控实名认证提升等级
          me.initCertError();
          me._setValidator();
          me._setEvent();
          me._checkCapsLock();
          me.setEventFooterBar();
          if (me.config.smsAutoLogin.mobile) {
            document.getElementById(me.$getId("smsPhone")).value = me.config.smsAutoLogin.mobile;
          }
        }
        // 判读是否是弹层登录，如果是弹层，则不在此进行请求
        if (me.config.diaPassLogin) {
          return handleRender();
        } else {
          // 客户端互通pc登录数据初始化
          return me._initClientShare(handleRender);
        }
      },
      _initApi: function (callbacks) {
        var me = this;
        me.initialized = true;
        me.initTime = new Date().getTime();
        me._getRSA(function (result) {
          me.RSA = result.RSA;
          me.rsakey = result.rsakey;
        });
        passport.data
          .getApiInfo({
            apiType: "login",
            gid: me.guideRandom || "",
            loginVersion: "v5",
            loginType: me.config && me.config.diaPassLogin ? "dialogLogin" : "basicLogin",
          })
          .success(function (rsp) {
            /**
             * @description 获取api初始化信息
             * @name magic.passport#getApiInfo
             * @grammar magic.passport.login#getApiInfo(e)
             * @event
             * @param {Object} e 事件参数
             * @param {Object} e.rsp 服务器返回信息
             */
            var returnValue = me.fireEvent("getApiInfo", {
              rsp: rsp,
            });
            if (!returnValue) {
              return;
            }

            if (rsp.data.disable == 1) {
              me.setGeneralError(me.lang.sysUpdate);
            }

            if (rsp.errInfo.no == 0) {
              var token = rsp.data.token;
              me.bdPsWtoken = rsp.data.token;

              // 获取登录历史记录
              me.loginrecord = {};
              if (!me.config.autosuggest) {
                // 如果是合并登录，而且开启了autosuggest
                if (me.config.memberPass && !me.constant.SUBMITFLAG) {
                  me.getElement("userName").value = rsp.data.rememberedUserName;
                }
              }
              var usagent = navigator.userAgent;
              var iosVersion = !navigator.userAgent.match(/OS [1-8]_\d[_\d]* like Mac OS X/i); // 为true则是ios9以上
              var isIos = !!navigator.userAgent.toString().match(/\(i[^;]+;( U;)? CPU.+Mac OS X/);
              var isPad = navigator.userAgent.toString().indexOf("iPad");
              if (iosVersion && isIos && isPad != null) {
                var a = document.getElementsByClassName("popBox");
                if (a != null && a.length > 0) {
                  if (window.screen.height > document.body.clientHeight) {
                    a[0].style.height =
                      window.screen.height * (window.screen.height / document.body.clientHeight) + 120 + "px";
                  } else {
                    a[0].style.height = window.screen.height * (window.screen.height / document.body.clientHeight);
                  }
                }
              }
              me.disUnameLogin = 0;
              me.ifShowWarning = rsp.data.ifShowWarning;
              if (rsp.data.spLogin && me.config.diaPassLogin) {
                me.spLogin = rsp.data.spLogin;
              }
              passport.data.setContext({
                token: token,
              });

              if (!navigator.cookieEnabled) {
                me.setGeneralError(me.lang.cookieDisable);
              }

              if (me.constant.SUBMITFLAG) {
                me.getElement("submit").click();
              } else {
                callbacks && callbacks.success(rsp);
              }
            }
          });
      },
      initPreUpsms: function () {
        var me = this;
        var upsms_dialog_tpl = me._getIrregularField("upsms_dialog");
        document.getElementsByTagName("body")[0].appendChild(baidu(upsms_dialog_tpl).get(0));
        var upsms = baidu("#" + me.$getId("upsms-dialog-btn"));
        baidu(me.getElement("upsms-dialog-wrap")).hide();
        upsms.on("click", function (e) {
          baidu(me.getElement("upsms-dialog-wrap")).hide();
          me._ownerDialog && me._ownerDialog.show();
        });
      },
      // 物联网手机号解禁
      initUnbindingSMS: function () {
        var me = this;
        try {
          var unbindingSMSTipsBtn = baidu("#" + me.$getId("unbindingSMSTipsBtn"));
          var unbindingSMSTipsBtn2 = baidu("#" + me.$getId("unbindingSMSTipsBtn2"));
          var TypeOptions = {
            goUnbinding: {
              continueTxt: "已知晓,去解禁",
              body:
                '<div class="unbinding-dialog-tip">' +
                "您百度账号绑定的手机号，是物联网卡手机号，根据相关法律法规要求，解禁后需要您在72小时内修改手机号，超时未修改的账号将会被关停" +
                "<br>操作路径：百度系APP→完成登录→我的→设置→账号管理→绑定手机→换绑手机号" +
                "</div>",
            },
            overUnbinding: {
              continueTxt: "联系百度客服",
              body:
                '<div class="unbinding-dialog-tip">' +
                "您的自助解禁次数已达上限，请联系百度官方客服进行人工解禁~" +
                "</div>",
            },
          };
          var unbindingSMSTipsBtnHandle = function (e) {
            var curOption = TypeOptions[me.config.curOption || "goUnbinding"];
            passport._load(me._domain.auto + "/passApi/js/uni_wrapper.js", true, function () {
              var toastWidget = passport.pop.init({
                type: "confirmWidget",
                tangram: true,
                titleText: "提示",
                width: 400,
                apiOpt: {
                  Cancel: "取消",
                  Continue: curOption.continueTxt,
                  contentHTML: curOption.body,
                },
                onConfirmClose: function () {
                  me._ownerDialog && me._ownerDialog.show();
                },
                onConfirmCancel: function () {
                  toastWidget.hide();
                  me._ownerDialog && me._ownerDialog.show();
                },
                onConfirmContinue: function () {
                  if (me.config.curOption === "goUnbinding") {
                    var data = me.config.curOptionDatas;
                    data.client = "pc";
                    data.clientfrom = "pc";
                    passport.data.unbindingSms(data).success(function (rsp) {
                      // console.log('rsp====>', rsp);
                      toastWidget.hide();
                      me.getElement("unbindingSMSTipsWrapper") &&
                        baidu(me.getElement("unbindingSMSTipsWrapper")).hide();
                      me.getElement("unbindingSMSTipsWrapper2") &&
                        baidu(me.getElement("unbindingSMSTipsWrapper2")).hide();
                      me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).show();
                      // me.getElement('smsError') && baidu(me.getElement('smsError')).show();
                      if (+rsp.data.code === 110000) {
                        me.showTipsToast("解禁成功,请重新登录");
                        me._ownerDialog && me._ownerDialog.show();
                      } else {
                        me._ownerDialog && me._ownerDialog.show();
                        // me.setGeneralError("解禁失败,请重试(${rsp.data.code})");
                        // me._setSmsGeneralError("解禁失败,请重试(${rsp.data.code})");
                        me.showTipsToast("用户当前状态非物联网卡封禁,请重新尝试(" + rsp.data.code + ")");
                      }
                    });
                  } else {
                    window.open(me.PcDukeUrl);
                  }
                },
              });
              toastWidget.show();
              me._ownerDialog && me._ownerDialog.hide();
            });
          };
          unbindingSMSTipsBtn && unbindingSMSTipsBtn.on("click", unbindingSMSTipsBtnHandle);
          unbindingSMSTipsBtn2 && unbindingSMSTipsBtn2.on("click", unbindingSMSTipsBtnHandle);
        } catch (error) {
          // console.log('initUnbindingSMS err2:', error);
        }
      },
      // 实名提升涉诈解禁
      initCertError: function () {
        var me = this;
        try {
          var certErrorBtn = baidu("#" + me.$getId("certErrorBtn"));
          var certErrorBtn2 = baidu("#" + me.$getId("certErrorBtn2"));
          var TypeOptions = {
            continueTxt: "已知晓,去解禁",
            body:
              '<div class="unbinding-dialog-tip">' +
              "您百度账号存在安全风险，解禁后需要您在48小时内完成实名认证，超时未完成认证的账号将会被限制登录" +
              "<br>操作路径：百度系APP→完成登录→我的→设置→帐号管理→身份认证→完成实名验证" +
              "</div>",
          };
          var certErrorBtnHandle = function (e) {
            passport._load(me._domain.auto + "/passApi/js/uni_wrapper.js", true, function () {
              var toastWidget = passport.pop.init({
                type: "confirmWidget",
                tangram: true,
                titleText: "提示",
                width: 400,
                apiOpt: {
                  Cancel: "取消",
                  Continue: TypeOptions.continueTxt,
                  contentHTML: TypeOptions.body,
                },
                onConfirmClose: function () {
                  me._ownerDialog && me._ownerDialog.show();
                },
                onConfirmCancel: function () {
                  toastWidget.hide();
                  me._ownerDialog && me._ownerDialog.show();
                },
                onConfirmContinue: function () {
                  var data = me.config.certOptions;
                  data.client = "pc";
                  data.clientfrom = "pc";
                  data.version = 2;
                  passport.data.certUpLevel(data).success(function (rsp) {
                    toastWidget.hide();
                    me.getElement("certErrorWrapper") && baidu(me.getElement("certErrorWrapper")).hide();
                    me.getElement("certErrorWrapper2") && baidu(me.getElement("certErrorWrapper2")).hide();
                    me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).show();

                    if (+rsp.data.code === 110000) {
                      me.showTipsToast("解禁成功,请重新登录");
                      me._ownerDialog && me._ownerDialog.show();
                    } else {
                      me._ownerDialog && me._ownerDialog.show();
                      me.showTipsToast("您的自助解禁次数已达上限，请联系百度官方客服进行人工解禁~");
                    }
                  });
                },
              });
              toastWidget.show();
              me._ownerDialog && me._ownerDialog.hide();
            });
          };
          certErrorBtn && certErrorBtn.on("click", certErrorBtnHandle);
          certErrorBtn2 && certErrorBtn2.on("click", certErrorBtnHandle);
        } catch (error) {
          // console.log('initcertError err2:', error);
        }
      },
      // thirdCheckAgreement=1时，弹窗确认协议（参照upsms-dialog-wrap实现）
      _showProtocalConfirmDialog: function (agreeCallback) {
        var me = this;
        var protocalUrl = me.constant.PROTOCAL_URL;
        var personalUrl = me.constant.PERSONAL_PROTOCAL_URL;
        var protocalText = me.config.protocalText || me.lang.baiduUserProtocal;
        var personalText = me.config.userpersonalText || me.lang.baiduUserPersonal;
        // 构建协议链接，与tab下的协议内容保持一致（含儿童协议和自定义协议）
        var agreementLinks =
          '<a target="_blank" href="' +
          protocalUrl +
          '">' +
          protocalText +
          "</a>" +
          " " +
          (me.customizeAnd ? "、" : me.lang.and) +
          " " +
          '<a target="_blank" href="' +
          personalUrl +
          '">' +
          personalText +
          "</a>";
        if (me.config.showChildAgree) {
          agreementLinks +=
            '、<a target="_blank" href="' + me.constant.CHILD_AGREEMENT_URL + '">' + me.lang.childAgreement + "</a>";
        }
        var customList = me.config.customAgreementList || [];
        for (var i = 0; i < customList.length; i++) {
          agreementLinks += '、<a target="_blank" href="' + customList[i].link + '">' + customList[i].text + "</a>";
        }
        var wrapId = me.$getId("protocalConfirmDialog");
        var wrap = document.getElementById(wrapId);
        if (!wrap) {
          var html =
            '<div class="pass-protocal-dialog-wrap" id="' +
            wrapId +
            '" style="display:none;">' +
            '<div class="pass-protocal-dialog-box">' +
            '<div class="pass-protocal-dialog-header">' +
            '<h5 class="pass-protocal-dialog-header-title">' +
            me.lang.thirdCheckDialogTitle +
            "</h5>" +
            "</div>" +
            '<div class="pass-protocal-dialog-body">' +
            '<p class="pass-protocal-dialog-body-txt">' +
            me.lang.thirdCheckDialogBody +
            "</p>" +
            '<p class="pass-protocal-dialog-body-txt">' +
            agreementLinks +
            "</p>" +
            "</div>" +
            '<div class="pass-protocal-dialog-footer">' +
            '<button class="pass-protocal-dialog-btn pass-protocal-dialog-btn-disagree" id="' +
            me.$getId("protocalDisagreeBtn") +
            '">' +
            me.lang.thirdCheckDialogDisagree +
            "</button>" +
            '<button class="pass-protocal-dialog-btn pass-protocal-dialog-btn-agree" id="' +
            me.$getId("protocalAgreeBtn") +
            '">' +
            me.lang.thirdCheckDialogAgree +
            "</button>" +
            "</div></div></div>";
          var temp = document.createElement("div");
          temp.innerHTML = html;
          wrap = temp.firstChild;
          document.body.appendChild(wrap);
        }
        var agreeBtn = document.getElementById(me.$getId("protocalAgreeBtn"));
        var disagreeBtn = document.getElementById(me.$getId("protocalDisagreeBtn"));
        if (agreeBtn) {
          agreeBtn.onclick = function () {
            wrap.style.display = "none";
            agreeCallback && agreeCallback();
          };
        }
        if (disagreeBtn) {
          disagreeBtn.onclick = function () {
            wrap.style.display = "none";
          };
        }
        wrap.style.display = "";
      },
      // 同步协议勾选状态到登录按钮（qrcode tab 无登录按钮不处理）
      _syncAgreeCheckboxState: function (currentTab, checked) {
        var me = this;
        if (currentTab === "qrcode") {
          return;
        }
        var submitBtn;
        if (currentTab === "sms") {
          submitBtn = me.getElement("smsSubmit");
        } else {
          submitBtn = me.getElement("submit");
        }
        if (submitBtn) {
          if (checked) {
            baidu.dom(submitBtn).removeClass("pass-button-full-disabled").attr("disabled", false);
          } else {
            baidu.dom(submitBtn).addClass("pass-button-full-disabled").attr("disabled", true);
          }
        }
      },
      showTipsToast: function (msg) {
        var me = this;
        var tipsToastEle = baidu("#" + me.$getId("tipToast-wrap"));
        var tipsToastBodyEle = baidu("#" + me.$getId("tipToast-body"));
        if (tipsToastEle && tipsToastBodyEle) {
          tipsToastEle.show();
          tipsToastBodyEle.addClass("tipToast-body-show");
          msg && (me.getElement("tipToast-body").innerText = msg);
          setTimeout(function () {
            tipsToastBodyEle.removeClass("tipToast-body-show");
            tipsToastEle.hide();
          }, 2000);
        }
      },
      submitForm: function () {
        /**
         * @description 提交当前表单
         * @name magic.passport.login#submitForm
         * @grammar magic.passport.login#submitForm
         * @function
         */
        var me = this;
        me.constant.SUBMITFLAG = true;
      },
      setSubpro: function (val) {
        var me = this;
        if (me.getElement("subpro") && val) {
          me.getElement("subpro").value = val
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/\x22/g, "&quot;")
            .replace(/\x27/g, "&#39;");
        }
      },
      _setValidator: function () {
        var me = this;
        if (!me.validatorInited) {
          // 用户名检测 邮箱长度
          me._validator.addRule("unameMailLength", function (ele) {
            var value = String(ele.value);
            if (/^[0-9a-zA-Z._-]+@([0-9a-zA-Z-]+\.)+[a-z]{2,4}$/.test(value)) {
              return value.length <= 60;
            }
            return true;
          });
          me._validator.addMsg("unameMailLength", me.lang.unameMailLengthError);
          me._validator.addRule("unameInputLogin", function (ele) {
            var value = String(ele.value);
            if (me.disUnameLogin == 0 && me.config.diaPassLogin) {
              if (!/^[0-9a-zA-Z._-]+@([0-9a-zA-Z-]+\.)+[a-z]{2,4}$/.test(value)) {
                return false;
              }
            }
            return true;
          });
          me._validator.addMsg("unameInputLogin", me.lang.unameInputError);
          me._validator.addRule("checkVcodeLength", function (ele, type) {
            var value = ele.value;
            if (!me.constant.CHECKVERIFYCODE) {
              me.$hide("verifyCodeSuccess");
              return false;
            }
            return true;
          });
          me._validator.addMsg("checkVcodeLength", me.lang.verifyCodeLenErr);
          me._validator.addRule("checkVcodeStatus", function (ele, type) {
            if (type == "all") {
              if (!me.constant.CHECKVERIFYCODE) {
                return false;
              }
            }
            return true;
          });
          me._validator.addMsg("checkVcodeStatus", me.lang.verifyCodeStaErr);
        }
        me.validatorInited = true;
        me.validateRules = {
          userName: {
            rules: ["required"],
            desc: me.lang.userName,
          },
          password: {
            rules: ["required"],
            desc: me.lang.password,
          },
          verifyCode: {
            rules: ["required", "checkVcodeLength", "checkVcodeStatus"],
            desc: me.lang.captcha,
          },
        };
        me._validator.init(me, me.validateRules);
      },
      _validateError: function (info, opt) {
        var me = this;
        var ele = baidu(me.getElement(info.field));
        ele.addClass(me.constant.ERROR_CLASS);
        if (me.config && me.config.hideEmailAndPhoneText && info.field == "userName") {
          me.setGeneralError(me.lang.enterUserName);
        } else {
          me.setGeneralError(info.msg);
        }
        if (me.disUnameLogin == 0 && info.field == "userName" && info.msg == me.lang.unameInputError) {
          var img = new Image();
          img.onload = img.onerror = function () {
            img.onload = img.onerror = null;
            img = null;
          };
          img.src = me._domain.https + "/img/v.gif?type=login&loginType=userName";
        }
        opt && opt.callback && callback();
      },
      _enableUnameLoginCallback: function (ele, type) {
        var me = this;
        var $usernameTemp = baidu('<input type="hidden" name="userNameLogin" value="1">');
        var placeholderDom = me.getElement("pass-pop-login-placeholder-normal");
        var container = me.getElement().parentNode;
        var tabLiDom = baidu(".tab li", container);
        var tabADom = baidu(".tab a", container).get(0);
        var $ele = baidu(ele);

        if (!ele && !type) {
          type = me.isLoginWeak == 1 ? "normal" : "other";
          $ele = me.eleLoginWeak;
        }
        if (me.disUnameLogin == 0) {
          me.disUnameLogin = 1;
          me._validator.confStorage[me.$getId()].userName.desc = me.lang.userNameTitle;
          $usernameTemp.get(0).value = "1";
          if (placeholderDom) {
            placeholderDom.innerHTML = me.lang.userNameTitle;
          }
          if (tabADom) {
            tabADom.innerHTML = me.lang.userNameLogin;
          }
          if (me.normalLogin) {
            me.normalLogin.innerHTML = me.lang.userNameLogin;
            baidu(me.normalLogin).addClass("pass-normal-btn-s2");
            if (me.normalLogin.style.display !== "none") {
              me._changeLoginType("normal");
            }
          }
          me.getElement("userNameLabel").innerHTML = me.lang.userNameTitle;
          var p = me.getElement("error");
          baidu(me.getElement("userName")).removeClass("pass-text-input-error");

          baidu(".tang-pass-pop-login-placeholder").hide();
          if (!me.getElement("userName").value) {
            me.$show("pass-pop-login-placeholder-normal");
          }

          me.changeSuggestView({
            list: "hide",
            btn: "hide",
          });
          if (type == "normal") {
            $ele.removeClass("pass-unamelogin-btn");
            $ele.addClass("pass-emaillogin-btn");
            $ele.get(0).innerHTML = me.lang.emailLogin;
          } else if (type == "other") {
            $ele.get(0).innerHTML =
              me.lang.forgetUserName +
              '<a href="###" id="pass-user-login" tabIndex="-1" data-click="other">' +
              me.lang.emailLogin +
              "</a>";
          }

          var img = new Image();
          img.onload = img.onerror = function () {
            img.onload = img.onerror = null;
            img = null;
          };
          img.src = me._domain.https + "/img/v.gif?type=login&loginType=normalName";
        } else {
          me.disUnameLogin = 0;
          me._validator.confStorage[me.$getId()].userName.desc = me.lang.account;
          $usernameTemp.get(0).value = "0";
          if (placeholderDom) {
            placeholderDom.innerHTML = me.lang.account;
          }
          if (tabADom) {
            tabADom.innerHTML = me.lang.emailLogin;
          }
          if (me.normalLogin) {
            me.normalLogin.innerHTML = me.lang.emailLogin;
            baidu(me.normalLogin).removeClass("pass-normal-btn-s2");
            if (me.normalLogin.style.display != "none") {
              me._changeLoginType("normal");
            }
          }

          me.getElement("userNameLabel").innerHTML = me.lang.account;
          var p = me.getElement("error");
          baidu(me.getElement("userName")).removeClass("pass-text-input-error");

          baidu(".tang-pass-pop-login-placeholder").hide();
          if (!me.getElement("userName").value) {
            baidu(me.getElement("pass-pop-login-placeholder-normal")).show();
          }
          me.changeSuggestView({
            list: "hide",
          });

          if (me.selectBtn && me.loginrecord && me.loginrecord.email && me.loginrecord.email.length > 1) {
            me.changeSuggestView({
              btn: "show",
            });
          } else if (me.selectBtn) {
            me.changeSuggestView({
              btn: "hide",
            });
          }
          if (type == "normal") {
            $ele.addClass("pass-unamelogin-btn");
            $ele.removeClass("pass-emaillogin-btn");
            $ele.get(0).innerHTML = me.lang.userNameLogin;
          } else if (type == "other") {
            // bca-disable-next-line
            $ele.get(0).innerHTML =
              me.lang.forgetEmail +
              '<a href="###" id="pass-user-login" tabIndex="-1" data-click="other">' +
              me.lang.userNameLogin +
              "</a>";
          }
        }
      },
      _validateSuccess: function (info, opt) {
        var me = this;
        var ele = baidu(me.getElement(info.field));
        me.clearGeneralError();
        ele.removeClass(me.constant.ERROR_CLASS);
        opt && opt.callback && callback();
      },
      _defaultLoginErr: function (rsp) {
        var me = this;
        me.vcodetype = rsp.data.vcodetype;
        // rsp.data.isslide = '1';
        if (rsp.data.isslide === "1") {
          rsp.errInfo["msg"] = "";
          me.loginPassMkd.initVcode({}, function () {
            me.getElement("submit").disabled = false;
            me.getElement("submit").value = me.lang.login;
            me.getElement("submit").style.color = "#fff";
          });
          me._ownerDialog && me._ownerDialog.hide("unHide");
        } else if (rsp.data.codeString) {
          me.getVerifyCode(rsp.data.codeString);
          me._clearInput("verifyCode");
          me.vcodefrom = "login";
        } else {
          me.clearVerifyCode();
        }
        // 新人机验证码命中
        if (+rsp.errInfo.no === 50052) {
          rsp.errInfo["msg"] = "";
          if (me.initMkdFlag) {
            me.loginPassMkdv2.show();
          } else {
            me.loginPassMkdv2.init();
          }
          me.getElement("submit").disabled = false;
          me.getElement("submit").value = me.lang.login;
          me.getElement("submit").style.color = "#fff";
          me._ownerDialog && me._ownerDialog.hide("unHide");
          return;
        }

        // 物联网手机号解禁
        if (+rsp.errInfo.no === 400605) {
          me.getElement("unbindingSMSTipsWrapper") && baidu(me.getElement("unbindingSMSTipsWrapper")).show();
          me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).hide();
          me.getElement("submit").value = me.lang.login;
          me.getElement("submit").disabled = false;
          // goUnbinding overUnbinding
          me.config.curOption = +rsp.data.hitWlwFreeTime ? "overUnbinding" : "goUnbinding";
          return;
        } else {
          me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).show();
        }

        // 实名提升涉诈解禁
        if (+rsp.errInfo.no === 402013) {
          me.config.certOptions = {
            ltoken: rsp.data.ltoken || "",
            lstr: rsp.data.lstr || "",
          };
          me.getElement("certErrorWrapper") && baidu(me.getElement("certErrorWrapper")).show();
          me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).hide();
          me.getElement("submit").value = me.lang.login;

          me.getElement("submit").disabled = false;
          return;
        } else {
          me.getElement("errorWrapper") && baidu(me.getElement("errorWrapper")).show();
        }

        if (rsp.errInfo.no == 400401) {
          // 需要用户选择输入的是用户名还是手机号
          if (me.getElement("choiceuser_article")) {
            me.$show("choiceuser_article");
          } else {
            var choiceTemplate = me._getIrregularField("choiceuser");
            me._insertAfterW(baidu(choiceTemplate).get(0), me.getElement());
            me._setEventChoiceUser();
          }
          baidu(me.getElement()).hide();
        }
        if (rsp.errInfo.no == 257) {
          // 登录检测需要填写密码，如果需要，则隐藏一部分选择用户
          baidu(me.getElement()).removeClass("tang-pass-login-hide");
          me.$show(me.getElement()).$hide("choiceuser_article");
        }
        if (rsp.errInfo.no == 6 || rsp.errInfo.no == 257) {
          // 如果是验证码错误，或者需要输入验证码，清空并聚焦验证码栏
          me._clearInput("verifyCode");
        }

        if (rsp.errInfo.no == 4) {
          // 如果是密码错误，清空密码栏和验证码栏
          me._clearInput("password");
          if (rsp.data.resetpwd) {
            var resetpwd = "";
            switch (rsp.data.resetpwd) {
              case "1":
                resetpwd = "1" + me.lang.passwordResetIn;
                break;
              case "2":
                resetpwd = "2" + me.lang.passwordResetIn;
                break;
              case "3":
                resetpwd = "3" + me.lang.passwordResetIn;
                break;
              case "4":
                resetpwd = "3" + me.lang.passwordResetOut;
                break;
            }
            if (resetpwd.length > 0) {
              rsp.errInfo.msg = me._format(me.lang.passwordResetWarn, {
                resetpwd: resetpwd,
              });
            }
          }
        }
        if (rsp.errInfo.no == 7) {
          var resetpwd = "";
          if (resetpwd.length > 0) {
            rsp.errInfo.msg = me._format(me.lang.passwordResetWarnNo, {
              resetpwd: resetpwd,
            });
          }
        }
        if (rsp.errInfo.msg && rsp.errInfo.msg.indexOf("#{") >= 0) {
          // 如果需要替换errrsp.errInfo.msg中的字符
          if (rsp.errInfo.no == 110024) {
            var linkUrl =
              me._domain.https +
              "/v2/?regnotify&action=resend&tpl=" +
              me.config.product +
              "&email=" +
              encodeURIComponent(rsp.data.mail) +
              "&u=" +
              encodeURIComponent(rsp.data.u);
            rsp.errInfo.msg = me._format(rsp.errInfo.msg, { gotourl: linkUrl });
          } else {
            var account = me
              .getElement("userName")
              .value.replace(/[\-\_\,\!\|\~\`\(\)\#\$\%\^\&\*\{\}\:\;\"\'\L\<\>\?]/g, "");
            rsp.errInfo.msg = me._format(rsp.errInfo.msg, {
              urldata: "&account=" + account + "&tpl=" + me.config.product + "&u=" + me.config.u,
            });
          }
        }

        if (rsp.errInfo["field"]) {
          // 展现错误信息
          me.setValidateError(rsp.errInfo["field"], rsp.errInfo["msg"], rsp);
        } else {
          me.setGeneralError(rsp.errInfo["msg"], rsp);
        }
        if (rsp.data.isslide !== "1") {
          var errNo = rsp && rsp.errInfo && rsp.errInfo.no;
          var authid = rsp && rsp.data && rsp.data.authid;
          var isRiskCheat = rsp && rsp.data && (rsp.data.secstate === "risk" || rsp.data.secstate === "cheat");
          // 新验证控件弹窗渲染完成放开按钮点击状态
          if (([5, 400031, 120019, 120021, 400023].includes(+errNo) || isRiskCheat) && authid) {
            return;
          }
          me.getElement("submit").disabled = false;
          me.getElement("submit").value = me.lang.login;
          me.getElement("submit").style.color = "#fff";
        }
      },
      _asyncValidate: {
        // 验证图形验证码是否正确
        checkVerifycode: function (callbacks, type) {
          // 异步校验，仅提供给输入时校验使用
          var me = this;
          var eleVcode = me.getElement("verifyCode");
          var eleCodeStr = me.getElement("codeString");
          passport.data
            .checkVerifycode({
              verifycode: me._SBCtoDBC(eleVcode.value),
              loginVersion: "v5",
              codestring: eleCodeStr.value,
            })
            .success(function (rsp) {
              /**
               * @description 用户名检查
               * @name magic.passport.reg#userNameCheck
               * @event
               * @grammar magic.passport.reg#userNameCheck(e)
               * @param {Object} e 事件参数
               * @param {String} e.rsp 服务器返回信息
               */
              var returnValue = me.fireEvent("checkVerifycode", {
                rsp: rsp,
              });
              if (!returnValue) {
                return;
              }

              if (+rsp.errInfo.no === 0) {
                callbacks && callbacks.success(rsp);
                me.$hide("verifyCode_clearbtn");
                // me.$show('verifyCodeSuccess');
                me.constant.CHECKVERIFYCODE = true;
              } else if (rsp.errInfo.no == 500002 || rsp.errInfo.no == 500018) {
                rsp.msg = rsp.errInfo.msg;
                callbacks && callbacks.error(rsp);
                me.$hide("verifyCodeSuccess");
                me.constant.CHECKVERIFYCODE = false;
              } else {
                callbacks && callbacks.success(rsp);
                me.$hide("verifyCodeSuccess");
                me.constant.CHECKVERIFYCODE = true;
              }
            });
        },
      },
      _format: function (source, opts) {
        source = String(source);
        var data = Array.prototype.slice.call(arguments, 1);
        var toString = Object.prototype.toString;
        if (data.length) {
          data =
            data.length === 1
              ? opts !== null && /\[object Array\]|\[object Object\]/.test(toString.call(opts))
                ? opts
                : data
              : data;
          return source.replace(/#\{(.+?)\}/g, function (match, key) {
            var replacer = data[key];
            // chrome 下 typeof /a/ == 'function'
            if ("[object Function]" === toString.call(replacer)) {
              replacer = replacer(key);
            }
            return "undefined" == typeof replacer ? "" : replacer;
          });
        }
        return source;
      },

      loginConnect: function (data, callback, normalLoginFn) {
        var me = this;
        var bindUrl = "/v2/unite-bind";
        var connectParams = {
          username: data.username,
          smsVcode: data.smsVcode || "",
          sms: data.sms || "",
        };
        var _buildQuery = function (query) {
          if (typeof query == "object") {
            var builder = [];
            for (var p in query) {
              var value = query[p];
              if (value !== undefined && value !== null) {
                if (builder.length) builder.push("&");
                var valueString = encodeURIComponent(
                  typeof value == "boolean" ? (value ? "1" : "0") : value.toString(),
                );
                builder.push(encodeURIComponent(p), "=", valueString);
              }
            }
            return builder.join("");
          }
          if (typeof query == "string") {
            return query;
          }
          return null;
        };

        var _appendQuery = function (url, query) {
          query = _buildQuery(query);
          if (typeof query == "string") {
            var hasQuery = /\?/g.test(url);
            url += (hasQuery ? "&" : "?") + _buildQuery(query);
          }
          return url;
        };
        normalLoginFn();
      },

      checkPhone: function (userid, phone) {
        var me = this;
        if (window.checkPhoneWidget) {
          if (window.checkPhoneWidget.setMakePhone) {
            window.checkPhoneWidget.setMakePhone(phone, userid);
          }
          me._ownerDialog && me._ownerDialog.hide("unHide");
          window.checkPhoneWidget.show();
        } else if (!window.checkPhoneExist) {
          passport._load(me._domain.auto + "/passApi/js/uni_wrapper.js", true, function () {
            window.checkPhoneWidget = passport.pop.init({
              type: "checkPhone",
              apiOpt: {
                u: me.config.u,
                product: me.config.product ? me.config.product : "",
                phone: phone,
                apiMargicInstance: me,
                token: me.bdPsWtoken,
                username: userid,
                isuserid: 1,
                noSynBdu: me.config.noSynBdu || "",
                staticPage: me.config.staticPage,
              },
              tangram: true,
              onHide: function () {
                me._ownerDialog && me._ownerDialog.show();
              },
            });
            me._ownerDialog && me._ownerDialog.hide("unHide");
            window.checkPhoneWidget.show();
          });
        }
      },

      _loginCheck: function (value, history) {
        // history 是否为历史记录中请求
        var me = this;
        var dvValue = document.getElementById("dv_Input")
          ? document.getElementById("dv_Input").value
          : (window.LG_DV_ARG && window.LG_DV_ARG.dvjsInput) || "";
        dvValue = dvValue.length > 1500 ? "" : dvValue;
        if (value.length) {
          passport.data
            .loginCheck({
              sub_source: "leadsetpwd",
              userName: value,
              loginVersion: "v5",
              dv: dvValue,
            })
            .success(function (rsp) {
              if (+rsp.errInfo.no === 0 && rsp.data.userid && !history) {
                me.checkPhone(rsp.data.userid, rsp.data.mobile);
                window.checkPhoneExist = true;
                me._ownerDialog && me._ownerDialog.hide("unHide");
              } else if (rsp.errInfo.no === "0" && rsp.data.isconnect) {
                me.setGeneralError(me.lang.nopassLead);
              } else if (rsp.data.codeString.length) {
                // 登录检测触发风控，出现验证码
                me.vcodetype = rsp.data.vcodetype;
                me.getVerifyCode(rsp.data.codeString);
                me.vcodefrom = "checkuname";
              } else {
                me.clearVerifyCode();
              }
            });
        }
      },

      //全角转半角
      _SBCtoDBC: function (str) {
        var result = "";
        if (str) {
          var len = str.length;
          for (var i = 0; i < len; i++) {
            var cCode = str.charCodeAt(i);
            // 全角与半角相差（除空格外）：65248（十进制）
            cCode = cCode >= 0xff01 && cCode <= 0xff5e ? cCode - 65248 : cCode;
            // 处理空格
            cCode = cCode == 0x03000 ? 0x0020 : cCode;
            result += String.fromCharCode(cCode);
          }
          return result;
        } else {
          return;
        }
      },

      storageSupport: function (run) {
        try {
          if ("localStorage" in window && window["localStorage"] !== null) {
            run();
          }
        } catch (e) {
          // catch
        }
      },
      // 提交按钮
      submit: function () {
        var me = this;
        // 如果 logincheck 太快，验证码栏展开此处 submit 不会被触发, 参见 submitNotTrigger
        if (me.internation && !me._validatorforeignmobileFn(me.getElement("foreignMobile"))) {
          return;
        }
        me.validateAll(
          {
            success: function () {
              me._doFocus("submit");
              me.submitStatus = 1;
              /**
               * @description 表单提交前
               * @name magic.passport.login#beforeSubmit
               * @event
               * @grammar magic.passport.login#beforeSubmit
               */
              var returnValue = me.fireEvent("beforeSubmit");
              if (!returnValue) {
                return;
              }
              me._logPass(me.urlData, {
                eventType: "pc-pwdLogin-submit",
              });
              me._callBehaviorCallback({
                type: "user-pwd-submit",
              });
              if (me.spLogin) {
                var usernameTemp = baidu('<input type="hidden" name="splogin" value="' + me.spLogin + '">');
                me.getElement("hiddenFields").appendChild(usernameTemp.get(0));
                me.spLogin = null;
              }

              var data = baidu.form.json(me.getElement("form"));
              data.token = me.bdPsWtoken;
              passport.data.setContext(baidu.extend({}, me.config));
              if (data.foreignusername) {
                data.foreignusername = me._SBCtoDBC(data.foreignusername);
              }
              data.userName = me._SBCtoDBC(data.userName);
              data.verifyCode = me._SBCtoDBC(data.verifyCode);
              var password_DBC = me._SBCtoDBC(me.getElement("password").value);
              if (me.RSA && me.rsakey) {
                var passwordVal = password_DBC;
                if (passwordVal.length < 128 && !me.config.safeFlag) {
                  data.password = baidu.url.escapeSymbol(me.RSA.encrypt(passwordVal));
                  data.rsakey = me.rsakey;
                  data.crypttype = 12;
                }
              }
              if (password_DBC && +me.config.mustVerify === 1) {
                data.mustVerify = 1;
              }
              /*
               ** 登录反馈，超时，出错处理
               */
              var submitEle = me.getElement("submit");
              var time = 15000;
              var submitInterval;
              me.getElement("submit").style.color = "#9ebef4";
              submitEle.value = me.lang.logining;
              me.getElement("submit").disabled = true;

              submitInterval = setTimeout(function () {
                if (me.submitStatus == 1 && !me.config.connect) {
                  me.setGeneralError(me.lang.submitTimeup);
                }
                me.getElement("submit").disabled = false;
                me.getElement("submit").style.color = "#fff";
                submitEle.value = me.lang.login;
              }, time);
              me.histroyData = data;

              /**
               * pc请求登录的最终函数
               */
              me.loginFn = function (mkdData, data) {
                // 初始化时间
                data.timeSpan = new Date().getTime() - me.initTime;
                if (passport.data.traceID) {
                  passport.data.traceID.startFlow && passport.data.traceID.startFlow("login");
                }
                // 请求登录接口
                if (me.internation) {
                  data.username = me._SBCtoDBC(me.getElement("foreignMobile").value);
                  data.isPhone = true;
                  data.countrycode = baidu(me.getElement("foreignMobileLabel")).attr("data-countrycode") || "";
                } else {
                  data.countrycode = "";
                }
                data.FP_UID = me._getCookie("FP_UID") || "";
                data.FP_INFO = window.PP_FP_INFO || "";
                data.loginVersion = "v5";
                data.supportdv = "1";
                data.bdint_sync_cookie = me.config.isSyncBaiduInt;

                if (mkdData && mkdData.ds && mkdData.tk) {
                  data.ds = mkdData.ds || "";
                  data.tk = mkdData.tk || "";
                }
                if (mkdData && mkdData.s && mkdData.k) {
                  data.s = mkdData.s || "";
                  data.k = mkdData.k || "";
                }
                data.dv = document.getElementById("dv_Input")
                  ? document.getElementById("dv_Input").value
                  : (window.LG_DV_ARG && window.LG_DV_ARG.dvjsInput) || "";
                // 添加验证码来源的参数用于统计
                if (me.vcodefrom) {
                  data.vcodefrom = me.vcodefrom;
                }
                data = baidu.extend(data, me.fuid || {});
                // TODO mzp
                passport.data.login(data).success(function (rsp) {
                  // 清除超时定时器
                  clearTimeout(submitInterval);
                  me.submitStatus = 2; // submitStatus = 2获取到登录验证信息
                  rsp.loginType = "password";
                  if (+rsp.errInfo.no === 0) {
                    /**
                     * @description 登录成功
                     * @name magic.passport.login#loginSuccess
                     * @event
                     * @grammar magic.passport.login#loginSuccess(e)
                     * @param {Object} e 事件参数
                     * @param {Object} e.data 服务器返回信息
                     * @param {Boolean} evt.returnValue 返回false时，阻止跳转
                     */
                    me.storageSupport(function () {
                      localStorage.setItem("passLoginType", "normal");
                    });
                    me.config.loginType = "pwdLogin";

                    var returnValue = me.fireEvent("loginSuccess", {
                      rsp: Object.assign({}, { type: "pwdLogin" }, rsp),
                    });

                    if (!returnValue) {
                      return;
                    }
                    if (window.location) {
                      window.location.href = rsp.data.u;
                    } else {
                      document.location.href = rsp.data.u;
                    }
                  } else {
                    /**
                     * @description 登录失败
                     * @name magic.passport.login#loginError
                     * @event
                     * @grammar magic.passport.login#loginError(e)
                     * @param {Object} e 事件参数
                     * @param {Object} e.rsp 服务器返回信息
                     */
                    // submitEle.value = me.lang.login;
                    // me.getElement('submit').style.color = '#fff';
                    var returnValue = me.fireEvent("loginError", {
                      rsp: Object.assign({}, { type: "pwdLogin" }, rsp),
                    });
                    me._logPass(me.urlData, {
                      eventType: "pc-pwdLogin-submit-error",
                      errno: rsp.errInfo.no,
                    });

                    if (!returnValue) {
                      return;
                    }
                    // 如果是i标记用户跳申诉
                    if (+rsp.errInfo.no === 500072 && !me.isEnVersion) {
                      me.getElement("submit").disabled = false;
                      me.getElement("submit").value = me.lang.login;
                      var targetjumpUrl = rsp.data.appealurl || "https://passport.baidu.com/?getpassappeal";
                      document.location.href = targetjumpUrl;
                      return;
                    }

                    if (+rsp.errInfo.no === 500072 && me.isEnVersion) {
                      me.getElement("submit").disabled = false;
                      me.getElement("submit").value = me.lang.login;
                      // bca-disable-next-line
                      me.getElement("error").innerHTML = "There is a security risk cannot login";
                      return;
                    }
                    // 保存参数用于解禁接口
                    me.config.curOptionDatas = data || {};
                    me.config.curOptionDatas.ltoken = rsp.data.ltoken || "";
                    me.config.curOptionDatas.lstr = rsp.data.lstr || "";
                    me._defaultLoginErr(rsp);
                  }
                });
              };
              // 新增一个方法 为了获取ds，tk
              /**
               * pc账号名密码登录回调函数 -获取人机数据提交给服务端
               */
              me.mkdDataLoginFn = function () {
                if (me.loginPassMkd && me.loginPassMkd.getDataAsync) {
                  me.currentPath = "mkdDataLoginFn";
                  var mkdData = {};
                  me.loginPassMkd.getDataAsync(function (rsp) {
                    mkdData.ds = rsp.ds || "xxx_loginv4";
                    mkdData.tk = rsp.tk || "xxx_loginv4";
                    if (me.loginPassMkdv2) {
                      me.loginPassMkdv2.getDataAsync().then(function (res) {
                        mkdData.s = res.ds || "";
                        mkdData.k = res.tk || "";
                        me.loginFn(mkdData, data);
                      });
                    } else {
                      me.loginFn(mkdData, data);
                    }
                    me.loginFn(mkdData, data);
                  });
                } else {
                  me.loginFn(null, data);
                }
              };
              me.loginConnect(
                {
                  username: data.userName,
                  password: data.password,
                },
                {
                  success: function () {
                    clearTimeout(submitInterval);
                    submitEle.value = me.lang.login;
                  },
                  fail: function (msg) {
                    clearTimeout(submitInterval);
                    submitEle.value = me.lang.login;
                    me.setGeneralError(msg);
                  },
                },
                me.mkdDataLoginFn,
              );
            },
          },
          true,
        );
      },
      _eventHandler: (function () {
        var me;
        var style = {
            focus: function (field, e) {
              /**
               * @description 表单域获得焦点
               * @name magic.passport.reg#fieldFocus
               * @event
               * @grammar magic.passport.login#fieldFocus(e)
               * @param {Object} e 事件参数
               * @param {TangramDOM} e.ele 触发 focus 事件的表单域
               */
              var returnValue = me.fireEvent("fieldFocus", {
                ele: this,
              });
              if (!returnValue) {
                return;
              }
              this.addClass(me.constant.FOCUS_CLASS);
              this.removeClass(me.constant.ERROR_CLASS);
            },
            blur: function (field, e) {
              /**
               * @description 表单域失去焦点
               * @name magic.passport.reg#fieldBlur
               * @event
               * @grammar magic.passport.login#fieldBlur(e)
               * @param {Object} e 事件参数
               * @param {TangramDOM} e.ele 触发 blur 事件的表单域
               */
              // 输入去空格
              if (me.getElement(field)) {
                me.getElement(field).value = me.getElement(field).value.replace(/\s/g, "");
              }
              var returnValue = me.fireEvent("fieldBlur", {
                ele: this,
              });
              if (!returnValue) {
                return;
              }
              this.removeClass(me.constant.FOCUS_CLASS);
            },
            mouseover: function (field, e) {
              /**
               * @description 鼠标移入表单域
               * @name magic.passport.reg#fieldMouseover
               * @event
               * @grammar magic.passport.login#fieldMouseover(e)
               * @param {Object} e 事件参数
               * @param {TangramDOM} e.ele 触发 mouseover 事件的表单域
               */
              var returnValue = me.fireEvent("fieldMouseover", {
                ele: this,
              });
              if (!returnValue) {
                return;
              }
              this.addClass(me.constant.HOVER_CLASS);
            },
            mouseout: function (field, e) {
              /**
               * @description 鼠标移出表单域
               * @name magic.passport.reg#fieldMouseout
               * @event
               * @grammar magic.passport.login#fieldMouseout(e)
               * @param {Object} e 事件参数
               * @param {TangramDOM} e.ele 触发 fieldMouseout 事件的表单域
               */
              var returnValue = me.fireEvent("fieldMouseout", {
                ele: this,
              });
              if (!returnValue) {
                return;
              }
              this.removeClass(me.constant.HOVER_CLASS);
            },
            keyup: function (field, e) {
              /**
               * @description 输入表单的时候
               * @name magic.passport.login#fieldKeyup
               * @event
               * @grammar magic.passport.login#fieldKeyup(e)
               * @param {Object} e 事件参数
               * @param {TangramDOM} e.ele 触发 blur 事件的表单域
               */
              var returnValue = me.fireEvent("fieldKeyup", {
                ele: this,
              });
              if (!returnValue) {
                return;
              }
            },
          },
          behaviour = {
            focus: {
              userName: function () {
                if (me.getElement("loginMerge")) {
                  me.getElement("loginMerge").value = "true";
                  me.getElement("isPhone").value = "";
                }
              },
              password: function (field, e) {
                if (
                  me.currentLoginType === "normal" ||
                  (me.config.composeTemplate && me.currentLoginType === "qrcode")
                ) {
                  me._getRSA(function (result) {
                    me.RSA = result.RSA;
                    me.rsakey = result.rsakey;
                  });
                }
              },
              verifyCode: function () {
                me.$hide("verifyCodeSuccess");
              },
            },
            blur: {
              userName: function (field, e) {},
              password: function (field, e) {
                var value = this.get(0).value;
                if (value.length) {
                  me.validate(field);
                }
              },
              verifyCode: function (field, e) {
                var value = this.get(0).value;
                if (value.length) {
                  me.validate(field);
                }
                var verifyCodeEle = me.getElement("verifyCode");
                var $verifyCodeEle = baidu(verifyCodeEle);
                if (verifyCodeEle.value) {
                  me._asyncValidate.checkVerifycode.call(me, {
                    error: function (rsp) {
                      $verifyCodeEle.addClass(me.constant.ERROR_CLASS);
                      me.setGeneralError(rsp.msg);
                    },
                    success: function () {
                      $verifyCodeEle.removeClass(me.constant.ERROR_CLASS);
                      me.clearGeneralError();
                    },
                  });
                } else {
                  me.$hide("verifyCodeSuccess");
                }
              },
            },
            change: {
              userName: function (field, e) {
                var value = this.get(0).value;
                me._loginCheck(value);
              },
              verifyCode: function (field, e) {},
            },
            click: {
              verifyCodeChange: function (field, e) {
                // 更换验证码
                me.getElement("verifyCode").value = "";
                me._doFocus("verifyCode");
                me.getVerifyCode();
                e.preventDefault();
              },
            },
            keyup: {},
            submit: function (e) {
              me.submit();
              e.preventDefault();
            },
          };
        return {
          entrance: function (e) {
            me = this;
            var target = baidu(e.target);
            var field = e.target.name;
            if (!field && e.target.id) {
              var matches = e.target.id.match(/\d+__(.*)$/);
              if (matches) {
                field = matches[1];
              }
            }
            if (!field) {
              return;
            }
            if (style.hasOwnProperty(e.type)) {
              style[e.type].apply(baidu(e.target), [field, e]);
            }
            if (behaviour.hasOwnProperty(e.type)) {
              if (typeof behaviour[e.type] == "function") {
                behaviour[e.type].apply(baidu(e.target), [e]);
              }
              if (behaviour[e.type].hasOwnProperty(field)) {
                behaviour[e.type][field].apply(baidu(e.target), [field, e]);
              }
            }

            if (!me.initialized && e.type == "focus") {
              me._initApi();
            }
          },
        };
      })(),
      /**
       * @description 析构
       * @name magic.passport.login#$dispose
       * @function
       * @grammar magic.passport.login#$dispose()
       */
      $dispose: function () {
        var me = this;
        if (me.disposed) {
          return;
        }
        // 重置上行短信失败标记
        me._isUpSmsOver = false;
        baidu.dom(me.getElement()).removeClass(me.constant.CONTAINER_CLASS);
        me.getElement().removeChild(me.getElement("form"));
        magic.Base.prototype.$dispose.call(me);
      },
    });
  return magic;
});
