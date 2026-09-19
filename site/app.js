// tabs
document.querySelectorAll('.tabs button').forEach(function (btn) {
  btn.addEventListener('click', function () {
    document.querySelectorAll('.tabs button').forEach(function (b) { b.classList.remove('active'); });
    btn.classList.add('active');
    var map = { guess: 'tab-guess', api: 'tab-api', funcs: 'tab-funcs' };
    ['tab-guess', 'tab-api', 'tab-funcs'].forEach(function (id) {
      document.getElementById(id).parentElement.classList.toggle('hidden', id !== map[btn.dataset.tab]);
    });
    document.getElementById('tab-name').textContent =
      btn.dataset.tab === 'guess' ? 'guess.cx' : btn.dataset.tab === 'api' ? 'mini_api.cx' : 'funcs_demo.cx';
  });
});

// copy buttons
function textOf(key) {
  if (key === 'hero') return document.getElementById('hero-code').innerText;
  var active = document.querySelector('.tabs button.active');
  var id = active.dataset.tab === 'guess' ? 'tab-guess' : active.dataset.tab === 'api' ? 'tab-api' : 'tab-funcs';
  return document.getElementById(id).innerText;
}
document.querySelectorAll('.copy').forEach(function (btn) {
  btn.addEventListener('click', function () {
    var t = textOf(btn.dataset.copy);
    function done() { btn.textContent = 'скопировано ✓'; setTimeout(function () { btn.textContent = 'копировать'; }, 1200); }
    if (navigator.clipboard) navigator.clipboard.writeText(t).then(done, done);
    else {
      var ta = document.createElement('textarea');
      ta.value = t; document.body.appendChild(ta); ta.select();
      try { document.execCommand('copy'); } catch (e) {}
      document.body.removeChild(ta); done();
    }
  });
});

// typing effect in terminal
(function () {
  var el = document.getElementById('typed');
  var lines = ['codex.exe new hello.cx', 'codex.exe hello.cx', 'Привет! Первая игра уже сегодня →'];
  var li = 0, ci = 0, del = false;
  function tick() {
    var s = lines[li];
    el.textContent = s.slice(0, ci);
    if (!del) {
      ci++;
      if (ci > s.length) { del = true; return void setTimeout(tick, 1400); }
      setTimeout(tick, 45);
    } else {
      ci--;
      if (ci < 0) { ci = 0; del = false; li = (li + 1) % lines.length; setTimeout(tick, 350); return; }
      setTimeout(tick, 18);
    }
  }
  tick();
  document.getElementById('year').textContent = new Date().getFullYear();
})();

// scroll reveal
(function () {
  var io = new IntersectionObserver(function (entries) {
    entries.forEach(function (e) {
      if (e.isIntersecting) { e.target.classList.add('vis'); io.unobserve(e.target); }
    });
  }, { threshold: 0.12 });
  document.querySelectorAll('.reveal').forEach(function (el) { io.observe(el); });
})();
