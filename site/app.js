document.querySelectorAll('.tabs button').forEach(function (btn) {
  btn.addEventListener('click', function () {
    document.querySelectorAll('.tabs button').forEach(function (b) { b.classList.remove('active'); });
    btn.classList.add('active');
    ['guess', 'api', 'funcs'].forEach(function (name) {
      document.getElementById('tab-' + name).classList.toggle('hidden', name !== btn.dataset.tab);
    });
  });
});
