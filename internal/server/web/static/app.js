// Fixes tab key in textarea
document.querySelector('textarea')?.addEventListener('keydown', function (e) {
  if (e.key.toLowerCase() === 'tab') {
    e.preventDefault();

    const start = this.selectionStart;
    const end = this.selectionEnd;

    // Set textarea value to: text before caret + tab + text after caret
    this.value = this.value.substring(0, start) + '\t' + this.value.substring(end);

    // Move caret to right position
    this.selectionStart = this.selectionEnd = start + 1;
  }
});

function switchFont(to) {
  const main = document.querySelector('.wysiwyg');

  if (to === 'sans') {
    main.classList.remove('font-serif', 'font-sans');
    main.classList.add('font-sans');

    document.querySelector('#serif').classList.remove('active');
    document.querySelector('#sans').classList.add('active');
  } else if (to === 'serif') {
    main.classList.remove('font-serif', 'font-sans');
    main.classList.add('font-serif');

    document.querySelector('#sans').classList.remove('active');
    document.querySelector('#serif').classList.add('active');
  }
}
