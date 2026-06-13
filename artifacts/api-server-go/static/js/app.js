'use strict';

// ─── API ────────────────────────────────────────────────────────────────────
const api = {
  async req(method, path, body, isForm) {
    const opts = { method, credentials: 'include', headers: {} };
    if (isForm) { opts.body = body; }
    else if (body) { opts.headers['Content-Type'] = 'application/json'; opts.body = JSON.stringify(body); }
    const res = await fetch('/api' + path, opts);
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || 'Ошибка запроса');
    return data;
  },
  get: (p) => api.req('GET', p),
  post: (p, b) => api.req('POST', p, b),
  put: (p, b) => api.req('PUT', p, b),
  patch: (p, b) => api.req('PATCH', p, b),
  del: (p) => api.req('DELETE', p),
  upload: (path, formData) => api.req('POST', path, formData, true),
};

// ─── STATE ──────────────────────────────────────────────────────────────────
let state = { user: null, route: '', routeParam: null };

// ─── ROUTER ─────────────────────────────────────────────────────────────────
function getRoute() {
  const hash = location.hash.replace('#', '') || '/';
  const m = hash.match(/^(\/[^/]*)(?:\/(.+))?/);
  return { path: m ? m[1] : '/', param: m ? m[2] : null };
}
function navigate(path) { location.hash = path; }
window.addEventListener('hashchange', () => {
  const { path, param } = getRoute();
  state.route = path; state.routeParam = param;
  render();
});

// ─── TOAST ──────────────────────────────────────────────────────────────────
function toast(msg, type = 'success') {
  const el = document.createElement('div');
  el.className = `toast ${type}`; el.textContent = msg;
  document.getElementById('toast-container').appendChild(el);
  setTimeout(() => el.remove(), 3200);
}

// ─── ICONS ──────────────────────────────────────────────────────────────────
const ic = {
  menu:    `<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>`,
  close:   `<svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
  home:    `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`,
  book:    `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>`,
  user:    `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`,
  shield:  `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`,
  check:   `<svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><polyline points="20 6 9 17 4 12"/></svg>`,
  back:    `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="15 18 9 12 15 6"/></svg>`,
  clock:   `<svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
  plus:    `<svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>`,
  edit:    `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>`,
  trash:   `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4h6v2"/></svg>`,
  star:    `<svg width="14" height="14" fill="currentColor" viewBox="0 0 24 24"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>`,
  camera:  `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>`,
  award:   `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="8" r="6"/><path d="M15.477 12.89L17 22l-5-3-5 3 1.523-9.11"/></svg>`,
  layers:  `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>`,
  zap:     `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>`,
  globe:   `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>`,
  tag:     `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>`,
  users:   `<svg width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>`,
};

// ─── HELPERS ────────────────────────────────────────────────────────────────
function initials(name) { return (name||'?').split(' ').map(w=>w[0]).join('').toUpperCase().slice(0,2); }
function dur(min) { return min>=60?`${Math.floor(min/60)}ч${min%60>0?' '+min%60+'м':''}`:min+'м'; }
function esc(s) { return String(s??'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

// Градиентный фон-заглушка для курса (без эмодзи).
const thumbGradients = [
  'linear-gradient(135deg,#6366f1,#8b5cf6)',
  'linear-gradient(135deg,#3b82f6,#6366f1)',
  'linear-gradient(135deg,#14b8a6,#3b82f6)',
  'linear-gradient(135deg,#f59e0b,#ef4444)',
  'linear-gradient(135deg,#22c55e,#14b8a6)',
  'linear-gradient(135deg,#ec4899,#8b5cf6)',
];
function courseThumbBg(id) { return thumbGradients[(id||0) % thumbGradients.length]; }

// ─── MOBILE SIDEBAR ─────────────────────────────────────────────────────────
let sidebarOpen = false;
function toggleSidebar() {
  sidebarOpen = !sidebarOpen;
  document.getElementById('sidebar-overlay')?.classList.toggle('active', sidebarOpen);
  document.querySelector('.sidebar')?.classList.toggle('open', sidebarOpen);
}
function closeSidebarMobile() {
  sidebarOpen = false;
  document.getElementById('sidebar-overlay')?.classList.remove('active');
  document.querySelector('.sidebar')?.classList.remove('open');
}

// ─── SIDEBAR ────────────────────────────────────────────────────────────────
function renderSidebar() {
  const u = state.user, r = state.route;
  const nav = [
    { path: '/',        label: 'Обзор',          icon: ic.home },
    { path: '/courses', label: 'Курсы',           icon: ic.book },
    { path: '/profile', label: 'Профиль',         icon: ic.user },
    ...(u?.isAdmin ? [{ path: '/admin', label: 'Админ', icon: ic.shield }] : []),
  ];
  return `
    <div class="mobile-header">
      <div class="mobile-logo">Маркетинг<span>Про</span></div>
      <button class="mobile-menu-btn" onclick="toggleSidebar()" aria-label="Меню">${ic.menu}</button>
    </div>
    <div id="sidebar-overlay" onclick="closeSidebarMobile()"></div>
    <div class="sidebar">
      <div class="sidebar-logo-wrap">
        <div class="sidebar-logo">Маркетинг<span>Про</span></div>
        <button class="sidebar-close-btn" onclick="closeSidebarMobile()">${ic.close}</button>
      </div>
      <nav class="sidebar-nav">
        ${nav.map(n=>`
          <div class="nav-item ${r===n.path?'active':''}" onclick="navigate('${n.path}');closeSidebarMobile()">
            ${n.icon} ${esc(n.label)}
          </div>`).join('')}
      </nav>
      ${u?`
        <div class="sidebar-bottom">
          <div class="user-row">
            <div class="avatar">${u.avatarUrl?`<img src="${esc(u.avatarUrl)}" alt="">`:esc(initials(u.name))}</div>
            <div class="user-info">
              <div class="user-name">${esc(u.name)}</div>
              <div class="user-email">${esc(u.email)}</div>
            </div>
          </div>
          <button class="logout-btn" onclick="logout()">Выйти</button>
        </div>`:``}
    </div>
    <nav class="mobile-bottom-nav">
      ${nav.map(n=>`
        <div class="mobile-nav-item ${r===n.path?'active':''}" onclick="navigate('${n.path}')">
          <span class="mobile-nav-icon">${n.icon}</span>
          <span class="mobile-nav-label">${esc(n.label)}</span>
        </div>`).join('')}
    </nav>`;
}

// ─── AUTH ────────────────────────────────────────────────────────────────────
function renderLogin() {
  return `<div class="auth-wrap"><div class="auth-card">
    <div class="auth-logo">МаркетингПро</div>
    <div class="auth-subtitle">Платформа онлайн-курсов по маркетингу</div>
    <div class="auth-title">Войти в аккаунт</div>
    <form onsubmit="doLogin(event)">
      <div class="form-group"><label>Email</label>
        <input type="email" name="email" placeholder="you@example.com" required autocomplete="email"></div>
      <div class="form-group"><label>Пароль</label>
        <input type="password" name="password" placeholder="••••••••" required autocomplete="current-password"></div>
      <div id="auth-error" class="form-error"></div>
      <button type="submit" class="btn btn-primary btn-block" id="auth-btn">Войти</button>
    </form>
    <div class="form-link">Нет аккаунта? <a href="#/register">Зарегистрироваться</a></div>
  </div></div>`;
}
function renderRegister() {
  return `<div class="auth-wrap"><div class="auth-card">
    <div class="auth-logo">МаркетингПро</div>
    <div class="auth-subtitle">Платформа онлайн-курсов по маркетингу</div>
    <div class="auth-title">Создать аккаунт</div>
    <form onsubmit="doRegister(event)">
      <div class="form-group"><label>Имя</label>
        <input type="text" name="name" placeholder="Иван Иванов" required></div>
      <div class="form-group"><label>Email</label>
        <input type="email" name="email" placeholder="you@example.com" required autocomplete="email"></div>
      <div class="form-group"><label>Пароль</label>
        <input type="password" name="password" placeholder="Минимум 6 символов" required autocomplete="new-password"></div>
      <div id="auth-error" class="form-error"></div>
      <button type="submit" class="btn btn-primary btn-block">Зарегистрироваться</button>
    </form>
    <div class="form-link">Уже есть аккаунт? <a href="#/login">Войти</a></div>
  </div></div>`;
}

// ─── DASHBOARD ──────────────────────────────────────────────────────────────
let dashData = null;
async function loadDashboard() {
  try { dashData = await api.get('/dashboard'); render(); }
  catch(e) { console.error(e); }
}

function renderDashboard() {
  if (!dashData) { loadDashboard(); return loading(); }
  const { enrollments, continueWatching, featuredCourses } = dashData;

  const instructorMap = {};
  [...featuredCourses].forEach(c => {
    if (!instructorMap[c.authorName]) instructorMap[c.authorName] = { name: c.authorName, courses: 0 };
    instructorMap[c.authorName].courses++;
  });
  const instructors = Object.values(instructorMap).slice(0, 4);

  return `<div class="page">
    <div class="hero">
      <div class="hero-content">
        <h2>Развивайте навыки<br>в сфере маркетинга</h2>
        <p>Профессиональные курсы от ведущих экспертов — для начинающих специалистов и опытных маркетологов.</p>
        <button class="btn btn-hero" onclick="navigate('/courses')">Смотреть курсы</button>
      </div>
      <div class="hero-deco">
        <div class="hero-badge">SEO</div>
        <div class="hero-badge" style="margin-top:12px;opacity:0.7">SMM</div>
        <div class="hero-badge" style="margin-top:12px;opacity:0.5">Email</div>
      </div>
    </div>

    <div class="stats-row">
      <div class="card stat-card">
        <div class="stat-icon">${ic.book}</div>
        <div class="stat-value">${featuredCourses.length}+</div>
        <div class="stat-label">Курсов на платформе</div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon">${ic.users}</div>
        <div class="stat-value">${instructors.length}+</div>
        <div class="stat-label">Преподавателей</div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon">${ic.award}</div>
        <div class="stat-value">${enrollments.length}</div>
        <div class="stat-label">Ваших курсов</div>
      </div>
    </div>

    ${instructors.length > 0 ? `
    <div class="section-header"><div class="section-title">Преподаватели</div></div>
    <div class="grid4" style="margin-bottom:28px">
      ${instructors.map(ins => `
        <div class="card instructor-card">
          <div class="instructor-avatar-circle">${esc(initials(ins.name))}</div>
          <div class="instructor-name">${esc(ins.name)}</div>
          <div class="instructor-meta">${ins.courses} курс${ins.courses > 1 ? 'а' : ''}</div>
        </div>`).join('')}
    </div>` : ''}

    ${continueWatching.length > 0 ? `
    <div class="section-header">
      <div class="section-title">Продолжить обучение</div>
    </div>
    <div class="grid3" style="margin-bottom:28px">
      ${continueWatching.map(e=>`
        <div class="card course-card" onclick="navigate('/courses/${e.courseId}')">
          <div class="course-thumb" style="background:${courseThumbBg(e.courseId)}">
            ${e.course.imageUrl?`<img src="${esc(e.course.imageUrl)}" alt="">`:`<span class="thumb-letter">${esc((e.course.title||'?')[0])}</span>`}
          </div>
          <div class="course-card-body">
            <div class="course-title">${esc(e.course.title)}</div>
            <div class="course-author">${esc(e.course.authorName)}</div>
            <div class="progress-label">${e.completedLessons} из ${e.totalLessons} уроков</div>
            <div class="progress-bar"><div class="progress-fill" style="width:${Math.round((e.completedLessons/Math.max(1,e.totalLessons))*100)}%"></div></div>
          </div>
        </div>`).join('')}
    </div>` : ''}

    <div class="section-header">
      <div class="section-title">Рекомендуемые курсы</div>
      <button class="btn btn-sm btn-outline" onclick="navigate('/courses')">Все курсы</button>
    </div>
    <div class="grid3">${featuredCourses.slice(0,6).map(c=>renderCourseCard(c)).join('')}</div>

    <div class="about-block">
      <div class="about-header">
        <div class="about-title">О платформе МаркетингПро</div>
        <div class="about-tagline">Ваш путь к профессии маркетолога начинается здесь</div>
      </div>
      <div class="about-text">
        МаркетингПро — это современная образовательная платформа, созданная специально для специалистов
        в сфере маркетинга и рекламы. Мы объединили лучших практикующих экспертов отрасли и разработали
        систему обучения, которая даёт не просто теорию, а готовые инструменты для работы.
      </div>
      <div class="about-text">
        На платформе вы найдёте курсы по SEO-продвижению, SMM, контент-маркетингу, email-рассылкам,
        аналитике и многим другим направлениям. Каждый курс построен по принципу "от основ к практике":
        сначала вы получаете теоретическую базу, затем отрабатываете навыки на реальных задачах.
      </div>
      <div class="about-stats">
        <div class="about-stat"><div class="about-stat-val">500+</div><div class="about-stat-lbl">Студентов прошли курсы</div></div>
        <div class="about-stat"><div class="about-stat-val">6</div><div class="about-stat-lbl">Направлений маркетинга</div></div>
        <div class="about-stat"><div class="about-stat-val">95%</div><div class="about-stat-lbl">Довольных учеников</div></div>
        <div class="about-stat"><div class="about-stat-val">24/7</div><div class="about-stat-lbl">Доступ к материалам</div></div>
      </div>
      <div class="about-features">
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.layers}</div>
          <div><strong>Практические знания</strong><div class="feature-desc">Только реальные кейсы, инструменты и рабочие схемы — никакой воды</div></div>
        </div>
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.globe}</div>
          <div><strong>Удобный формат</strong><div class="feature-desc">Учитесь в любое время и в любом месте, в удобном для вас темпе</div></div>
        </div>
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.award}</div>
          <div><strong>Эксперты рынка</strong><div class="feature-desc">Курсы ведут практикующие специалисты с реальным опытом в отрасли</div></div>
        </div>
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.tag}</div>
          <div><strong>Бесплатный старт</strong><div class="feature-desc">Базовые курсы полностью бесплатны — начните обучение прямо сейчас</div></div>
        </div>
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.zap}</div>
          <div><strong>Быстрый результат</strong><div class="feature-desc">Уже после первых уроков вы сможете применять знания на практике</div></div>
        </div>
        <div class="about-feature">
          <div class="feature-icon-wrap">${ic.users}</div>
          <div><strong>Сообщество</strong><div class="feature-desc">Общайтесь с другими студентами и обменивайтесь опытом</div></div>
        </div>
      </div>
    </div>
  </div>`;
}

// ─── COURSES ────────────────────────────────────────────────────────────────
// allCoursesData — полный список всех курсов (никогда не фильтруется).
// coursesData — отфильтрованный список для отображения.
let allCoursesData = null, coursesData = null;
let courseFilter = { search:'', category:'', paid:'' };

async function loadCourses() {
  // Загружаем полный список только один раз.
  if (!allCoursesData) {
    try { allCoursesData = await api.get('/courses'); }
    catch(e) { console.error(e); return; }
  }
  applyFilters();
}

function applyFilters() {
  const { search, category, paid } = courseFilter;
  coursesData = (allCoursesData || []).filter(c => {
    if (search) {
      const q = search.toLowerCase();
      if (!c.title.toLowerCase().includes(q) && !c.description.toLowerCase().includes(q) && !c.authorName.toLowerCase().includes(q)) return false;
    }
    if (category && c.category !== category) return false;
    if (paid === 'true' && !c.isPaid) return false;
    if (paid === 'false' && c.isPaid) return false;
    return true;
  });
  render();
}

function renderCourseCard(c) {
  return `<div class="card course-card" onclick="navigate('/courses/${c.id}')">
    <div class="course-thumb" style="background:${courseThumbBg(c.id)}">
      ${c.imageUrl
        ? `<img src="${esc(c.imageUrl)}" alt="">`
        : `<span class="thumb-letter">${esc((c.title||'?')[0])}</span>`}
    </div>
    <div class="course-card-body">
      <div class="course-card-top">
        <span class="course-cat-tag">${esc(c.category)}</span>
        <span class="course-badge ${c.isPaid?'badge-paid':'badge-free'}">${c.isPaid?`${c.price} ₽`:'Бесплатно'}</span>
      </div>
      <div class="course-title">${esc(c.title)}</div>
      <div class="course-desc-preview">${esc(c.description).slice(0,90)}${c.description.length>90?'…':''}</div>
      <div class="course-card-footer">
        <div class="course-author">${esc(c.authorName)}</div>
        <div class="course-meta">${ic.clock} ${dur(c.duration)} &nbsp;·&nbsp; ${c.lessonsCount} ур.</div>
      </div>
    </div>
  </div>`;
}

function renderCourses() {
  if (!coursesData && !allCoursesData) { loadCourses(); return loading(); }
  if (!coursesData) { applyFilters(); return loading(); }

  // Все категории всегда берём из полного списка.
  const cats = [...new Set((allCoursesData||[]).map(c=>c.category))].filter(Boolean).sort();

  return `<div class="page">
    <div class="page-title">Каталог курсов</div>
    <div class="filter-row">
      <input class="search-input" type="text" placeholder="Поиск по названию, автору..." value="${esc(courseFilter.search)}"
        oninput="courseFilter.search=this.value;applyFilters()">
      <select class="filter-select" onchange="courseFilter.category=this.value;applyFilters()">
        <option value="">Все категории</option>
        ${cats.map(c=>`<option value="${esc(c)}" ${courseFilter.category===c?'selected':''}>${esc(c)}</option>`).join('')}
      </select>
      <select class="filter-select" onchange="courseFilter.paid=this.value;applyFilters()">
        <option value="">Все курсы</option>
        <option value="false" ${courseFilter.paid==='false'?'selected':''}>Бесплатные</option>
        <option value="true" ${courseFilter.paid==='true'?'selected':''}>Платные</option>
      </select>
    </div>
    ${courseFilter.category||courseFilter.paid||courseFilter.search ? `
    <div class="filter-active-row">
      ${courseFilter.category?`<span class="filter-chip">${esc(courseFilter.category)} <span onclick="courseFilter.category='';applyFilters()">×</span></span>`:''}
      ${courseFilter.paid==='true'?`<span class="filter-chip">Платные <span onclick="courseFilter.paid='';applyFilters()">×</span></span>`:''}
      ${courseFilter.paid==='false'?`<span class="filter-chip">Бесплатные <span onclick="courseFilter.paid='';applyFilters()">×</span></span>`:''}
      ${courseFilter.search?`<span class="filter-chip">«${esc(courseFilter.search)}» <span onclick="courseFilter.search='';applyFilters()">×</span></span>`:''}
      <button class="btn btn-sm btn-outline" style="margin-left:4px;padding:3px 10px;font-size:12px" onclick="courseFilter={search:'',category:'',paid:''};applyFilters()">Сбросить</button>
    </div>` : ''}
    <div class="courses-count" style="font-size:13px;color:var(--text2);margin-bottom:16px">Найдено курсов: ${coursesData.length}</div>
    ${coursesData.length===0
      ?`<div class="empty-state"><div class="empty-icon">${ic.book}</div><div>Курсы не найдены</div><div style="font-size:13px;color:var(--text2);margin-top:6px">Попробуйте изменить фильтры</div></div>`
      :`<div class="grid3">${coursesData.map(c=>renderCourseCard(c)).join('')}</div>`}
  </div>`;
}

// ─── COURSE DETAIL ──────────────────────────────────────────────────────────
let courseDetail = null, completedLessons = new Set();
async function loadCourseDetail(id) {
  try {
    const [course, enrollments] = await Promise.all([
      api.get('/courses/'+id),
      api.get('/enrollments').catch(()=>[]),
    ]);
    const enr = enrollments.find(e=>e.courseId===parseInt(id));
    courseDetail = { ...course, enrolled: !!enr };
    if (enr) {
      completedLessons = new Set();
    }
    render();
  } catch(e) { console.error(e); }
}
async function enroll(courseId) {
  try {
    await api.post('/enrollments', { courseId: parseInt(courseId) });
    toast('Вы записаны на курс!'); courseDetail=null; dashData=null; loadCourseDetail(courseId);
  } catch(e) { toast(e.message,'error'); }
}
async function toggleLesson(lessonId, courseId, done) {
  try {
    await api.post('/progress', { lessonId:parseInt(lessonId), courseId:parseInt(courseId), completed:done });
    if (done) completedLessons.add(lessonId); else completedLessons.delete(lessonId);
    dashData=null; render();
  } catch(e) { toast(e.message,'error'); }
}
function renderCourseDetail(id) {
  if (!courseDetail || courseDetail.id!==parseInt(id)) {
    courseDetail=null; completedLessons=new Set(); loadCourseDetail(id); return loading();
  }
  const c = courseDetail;
  const lessons = c.lessons || [];
  const doneCount = lessons.filter(l=>completedLessons.has(l.id)).length;
  const progress = lessons.length > 0 ? Math.round((doneCount / lessons.length) * 100) : 0;

  return `<div class="page">
    <div class="back-btn" onclick="navigate('/courses')">${ic.back} Назад к курсам</div>

    <div class="course-detail-hero" style="background:${courseThumbBg(c.id)}">
      ${c.imageUrl?`<img src="${esc(c.imageUrl)}" alt="" class="course-detail-hero-img">`:`
        <div class="course-detail-hero-inner">
          <div class="course-detail-letter">${esc((c.title||'?')[0])}</div>
          <div class="course-detail-cat">${esc(c.category)}</div>
        </div>`}
    </div>

    <div class="course-detail-layout">
      <div class="course-detail-main">
        <h1 class="course-detail-title">${esc(c.title)}</h1>
        <div class="course-detail-meta">
          <span class="detail-meta-item">${ic.users} ${esc(c.authorName)}</span>
          <span class="detail-meta-item">${ic.clock} ${dur(c.duration)}</span>
          <span class="detail-meta-item">${ic.book} ${c.lessonsCount} уроков</span>
          <span class="course-cat-tag">${esc(c.category)}</span>
        </div>

        <div class="course-detail-section">
          <div class="section-title">Описание курса</div>
          <p class="course-detail-desc">${esc(c.description)}</p>
        </div>

        ${c.enrolled && lessons.length > 0 ? `
        <div class="course-detail-section">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px">
            <div class="section-title">Ваш прогресс</div>
            <div style="font-size:13px;color:var(--text2)">${doneCount} из ${lessons.length} уроков (${progress}%)</div>
          </div>
          <div class="progress-bar" style="height:8px">
            <div class="progress-fill" style="width:${progress}%"></div>
          </div>
        </div>` : ''}

        <div class="course-detail-section">
          <div class="section-title" style="margin-bottom:16px">Программа курса</div>
          <div class="card">
            ${lessons.map((l,i)=>{
              const done = completedLessons.has(l.id);
              return `<div class="lesson-item">
                <div class="lesson-num ${done?'done':''}">${done?ic.check:i+1}</div>
                <div class="lesson-info">
                  <div class="lesson-title">${esc(l.title)}</div>
                  ${l.description?`<div class="lesson-desc">${esc(l.description)}</div>`:''}
                  <div class="lesson-meta">
                    ${ic.clock} ${dur(l.duration)}
                    ${l.isFree?`<span class="lesson-free-tag">Бесплатный урок</span>`:''}
                  </div>
                </div>
                ${c.enrolled?`<button class="btn btn-sm ${done?'btn-done':'btn-primary'}" onclick="toggleLesson(${l.id},${c.id},${!done})">${done?'Готово':'Пройти'}</button>`:''}
              </div>`;
            }).join('')}
            ${lessons.length===0?`<div style="padding:24px;color:var(--text2);font-size:14px;text-align:center">Уроки ещё добавляются</div>`:''}
          </div>
        </div>
      </div>

      <div class="course-detail-side">
        <div class="enroll-box">
          <div class="enroll-price ${!c.isPaid?'free':''}">${c.isPaid?(c.price+' ₽'):'Бесплатно'}</div>
          <div class="enroll-price-sub">${c.isPaid?'Полный доступ к курсу':'Бесплатный доступ'}</div>
          ${c.enrolled
            ?`<button class="btn btn-success btn-block" disabled>${ic.check} Вы записаны</button>`
            :`<button class="btn btn-primary btn-block" onclick="enroll(${c.id})">Записаться на курс</button>`}
          <div class="enroll-features">
            <div class="enroll-feature">${ic.check} ${c.lessonsCount} видеоуроков</div>
            <div class="enroll-feature">${ic.check} ${dur(c.duration)} обучения</div>
            <div class="enroll-feature">${ic.check} Доступ навсегда</div>
            ${!c.isPaid?`<div class="enroll-feature">${ic.check} Полностью бесплатно</div>`:''}
          </div>
        </div>

        <div class="card author-card">
          <div class="author-avatar">${esc(initials(c.authorName))}</div>
          <div class="author-name">${esc(c.authorName)}</div>
          <div class="author-role">Преподаватель курса</div>
        </div>
      </div>
    </div>
  </div>`;
}

// ─── PROFILE ────────────────────────────────────────────────────────────────
let profileProgress = null, profileEnrollments = null;
async function loadProfileProgress() {
  try {
    const [prog, enrs] = await Promise.all([
      api.get('/progress'),
      api.get('/enrollments').catch(()=>[]),
    ]);
    profileProgress = prog;
    profileEnrollments = enrs;
    render();
  } catch(e) { console.error(e); }
}

async function uploadAvatar(input) {
  if (!input.files[0]) return;
  const fd = new FormData();
  fd.append('file', input.files[0]);
  try {
    const data = await api.upload('/upload', fd);
    const urlField = document.getElementById('avatar-url-val');
    if (urlField) urlField.value = data.url;
    const preview = document.getElementById('avatar-preview');
    if (preview) preview.innerHTML = `<img src="${esc(data.url)}" alt="">`;
    toast('Фото загружено');
  } catch(e) { toast(e.message,'error'); }
}

function renderProfile() {
  const u = state.user;
  if (!profileProgress) { loadProfileProgress(); }
  const prog = profileProgress;
  const enrs = profileEnrollments || [];
  const maxBar = prog ? Math.max(1, ...prog.days.map(d=>d.lessonsCompleted)) : 1;

  // Общий прогресс по всем курсам
  const totalLessons = enrs.reduce((s,e)=>s+e.totalLessons, 0);
  const totalDone = enrs.reduce((s,e)=>s+e.completedLessons, 0);
  const overallPct = totalLessons > 0 ? Math.round((totalDone/totalLessons)*100) : 0;

  return `<div class="page">
    <div class="page-title">Профиль</div>
    <div class="profile-grid">
      <div class="card profile-avatar-wrap">
        <div class="profile-avatar" id="avatar-preview">
          ${u.avatarUrl?`<img src="${esc(u.avatarUrl)}" alt="">`:esc(initials(u.name))}
        </div>
        <div class="profile-name">${esc(u.name)}</div>
        <div class="profile-email">${esc(u.email)}</div>
        ${u.isAdmin?`<div style="margin-top:8px"><span class="course-badge badge-paid">${ic.shield} Администратор</span></div>`:''}
        <label class="upload-btn" style="margin-top:16px">
          ${ic.camera} Изменить фото
          <input type="file" accept="image/*" style="display:none" onchange="uploadAvatar(this)">
        </label>
      </div>
      <div class="card profile-form">
        <div class="section-title" style="margin-bottom:20px">Редактировать профиль</div>
        <form onsubmit="saveProfile(event)">
          <input type="hidden" id="avatar-url-val" value="${esc(u.avatarUrl||'')}">
          <div class="form-group"><label>Имя</label>
            <input type="text" name="name" value="${esc(u.name)}" required></div>
          <div class="form-group"><label>Дата рождения</label>
            <input type="date" name="dateOfBirth" value="${esc(u.dateOfBirth||'')}"></div>
          <div id="profile-error" class="form-error"></div>
          <button type="submit" class="btn btn-primary">Сохранить изменения</button>
        </form>
      </div>
    </div>

    ${enrs.length > 0 ? `
    <div class="section-title" style="margin:32px 0 14px">Мои курсы</div>
    <div class="card" style="margin-bottom:24px;overflow:visible">
      <div class="profile-progress-summary">
        <div>
          <div style="font-size:13px;color:var(--text2);margin-bottom:4px">Общий прогресс</div>
          <div style="font-size:22px;font-weight:700;color:var(--primary)">${overallPct}%</div>
        </div>
        <div>
          <div style="font-size:13px;color:var(--text2);margin-bottom:4px">Уроков пройдено</div>
          <div style="font-size:22px;font-weight:700">${totalDone} / ${totalLessons}</div>
        </div>
        <div>
          <div style="font-size:13px;color:var(--text2);margin-bottom:4px">Курсов записано</div>
          <div style="font-size:22px;font-weight:700">${enrs.length}</div>
        </div>
      </div>
      ${enrs.map(e => {
        const pct = e.totalLessons > 0 ? Math.round((e.completedLessons/e.totalLessons)*100) : 0;
        return `<div class="profile-course-row" onclick="navigate('/courses/${e.courseId}')">
          <div class="profile-course-thumb" style="background:${courseThumbBg(e.courseId)}">
            ${e.course.imageUrl?`<img src="${esc(e.course.imageUrl)}" alt="">`:esc((e.course.title||'?')[0])}
          </div>
          <div class="profile-course-info">
            <div class="profile-course-title">${esc(e.course.title)}</div>
            <div class="profile-course-meta">${esc(e.course.authorName)} &nbsp;·&nbsp; ${e.course.lessonsCount} уроков</div>
            <div class="profile-progress-row">
              <div class="progress-bar" style="flex:1;height:6px">
                <div class="progress-fill" style="width:${pct}%"></div>
              </div>
              <span class="profile-pct">${pct}%</span>
            </div>
            <div style="font-size:12px;color:var(--text2);margin-top:4px">${e.completedLessons} из ${e.totalLessons} уроков</div>
          </div>
          <div class="profile-course-pct-badge ${pct===100?'done':''}">${pct===100?ic.check:pct+'%'}</div>
        </div>`;
      }).join('')}
    </div>` : ''}

    <div class="section-title" style="margin:${enrs.length>0?'0':'32px'} 0 14px">Активность за неделю</div>
    ${prog ? `
    <div class="card" style="margin-bottom:24px">
      <div style="padding:16px 16px 0;display:flex;align-items:center;justify-content:space-between">
        <div style="font-size:14px;color:var(--text2)">Завершено уроков на этой неделе: <strong style="color:var(--text)">${prog.totalCompleted}</strong></div>
      </div>
      <div class="week-chart">
        ${prog.days.map(d=>`
          <div class="week-bar-wrap">
            <div class="week-count">${d.lessonsCompleted>0?d.lessonsCompleted:''}</div>
            <div class="week-bar-track">
              <div class="week-bar-fill" style="height:${d.lessonsCompleted===0?2:Math.round((d.lessonsCompleted/maxBar)*100)}%"></div>
            </div>
            <div class="week-bar-label">${esc(d.day)}</div>
          </div>`).join('')}
      </div>
      <div style="height:8px"></div>
    </div>` : `<div class="card" style="padding:24px;color:var(--text2);font-size:14px">Загрузка данных...</div>`}
  </div>`;
}

// ─── ADMIN ───────────────────────────────────────────────────────────────────
let adminCourses = null, adminEditId = null, showAdminForm = false;

async function loadAdminCourses() {
  try { adminCourses = await api.get('/admin/courses'); render(); }
  catch(e) { toast(e.message,'error'); }
}
async function adminBecomeAdmin(e) {
  e.preventDefault();
  const f = e.target;
  try {
    await api.post('/admin/become', { secret: f.secret.value });
    toast('Права администратора активированы!');
    state.user = await api.get('/auth/me');
    render();
  } catch(err) { toast(err.message,'error'); }
}
async function adminSaveCourse(e) {
  e.preventDefault();
  const f = e.target;
  const payload = {
    title: f.title.value,
    description: f.description.value,
    category: f.category.value,
    authorName: f.authorName.value,
    isPaid: f.isPaid.value === 'true',
    price: f.isPaid.value === 'true' ? parseFloat(f.price.value||0) : null,
    duration: parseInt(f.duration.value||0),
    lessonsCount: parseInt(f.lessonsCount.value||0),
    imageUrl: f.imageUrl.value || null,
  };
  try {
    if (adminEditId) {
      await api.put('/admin/courses/'+adminEditId, payload);
      toast('Курс обновлён');
    } else {
      await api.post('/admin/courses', payload);
      toast('Курс создан');
    }
    adminCourses=null; adminEditId=null; showAdminForm=false;
    allCoursesData=null; coursesData=null; dashData=null;
    loadAdminCourses();
  } catch(err) { toast(err.message,'error'); }
}
async function adminUploadCourseImage(input) {
  if (!input.files[0]) return;
  const fd = new FormData();
  fd.append('file', input.files[0]);
  try {
    const data = await api.upload('/upload', fd);
    const urlInput = document.getElementById('admin-img-url');
    if (urlInput) { urlInput.value = data.url; toast('Изображение загружено'); }
  } catch(e) { toast(e.message,'error'); }
}
async function adminDeleteCourse(id, title) {
  if (!confirm(`Удалить курс «${title}»?`)) return;
  try {
    await api.del('/admin/courses/'+id);
    toast('Курс удалён'); adminCourses=null; allCoursesData=null; coursesData=null; dashData=null; loadAdminCourses();
  } catch(e) { toast(e.message,'error'); }
}
function adminEditCourse(id) {
  adminEditId = id; showAdminForm = true; render();
  setTimeout(()=>{
    const c = adminCourses.find(x=>x.id===id);
    if (!c) return;
    const f = document.getElementById('admin-course-form');
    if (!f) return;
    f.title.value = c.title;
    f.description.value = c.description;
    f.category.value = c.category;
    f.authorName.value = c.authorName;
    f.isPaid.value = c.isPaid ? 'true' : 'false';
    f.price.value = c.price || '';
    f.duration.value = c.duration;
    f.lessonsCount.value = c.lessonsCount;
    f.imageUrl.value = c.imageUrl || '';
    togglePriceField();
  }, 0);
}
function togglePriceField() {
  const sel = document.getElementById('admin-ispaid');
  const row = document.getElementById('admin-price-row');
  if (sel && row) row.style.display = sel.value === 'true' ? 'block' : 'none';
}
function renderAdmin() {
  const u = state.user;
  if (!u.isAdmin) {
    return `<div class="page">
      <div class="page-title">Панель администратора</div>
      <div class="card" style="padding:32px;max-width:400px">
        <div class="section-title" style="margin-bottom:8px">Активация прав</div>
        <p style="color:var(--text2);font-size:14px;margin-bottom:20px">Введите секретный код для получения прав администратора.</p>
        <form onsubmit="adminBecomeAdmin(event)">
          <div class="form-group"><label>Секретный код</label>
            <input type="password" name="secret" placeholder="Введите код" required></div>
          <button type="submit" class="btn btn-primary">Активировать</button>
        </form>
      </div>
    </div>`;
  }
  if (!adminCourses) { loadAdminCourses(); return loading(); }
  return `<div class="page">
    <div class="page-title">${ic.shield} Панель администратора</div>

    ${showAdminForm ? `
    <div class="card admin-form-card">
      <div class="section-header" style="margin-bottom:16px">
        <div class="section-title">${adminEditId?'Редактировать курс':'Добавить новый курс'}</div>
        <button class="btn btn-sm btn-outline" onclick="adminEditId=null;showAdminForm=false;render()">Отмена</button>
      </div>
      <form id="admin-course-form" onsubmit="adminSaveCourse(event)">
        <div class="admin-form-grid">
          <div class="form-group"><label>Название *</label>
            <input type="text" name="title" required placeholder="Название курса"></div>
          <div class="form-group"><label>Преподаватель *</label>
            <input type="text" name="authorName" required placeholder="Имя преподавателя"></div>
          <div class="form-group"><label>Категория *</label>
            <input type="text" name="category" required placeholder="SEO, SMM, Контент..."></div>
          <div class="form-group"><label>Тип курса</label>
            <select name="isPaid" id="admin-ispaid" class="filter-select" style="width:100%" onchange="togglePriceField()">
              <option value="false">Бесплатный</option>
              <option value="true">Платный</option>
            </select></div>
          <div class="form-group" id="admin-price-row" style="display:none"><label>Цена (₽)</label>
            <input type="number" name="price" placeholder="2990" min="0"></div>
          <div class="form-group"><label>Длительность (минут)</label>
            <input type="number" name="duration" placeholder="360" min="0"></div>
          <div class="form-group"><label>Количество уроков</label>
            <input type="number" name="lessonsCount" placeholder="12" min="0"></div>
        </div>
        <div class="form-group"><label>Описание</label>
          <textarea name="description" rows="4" style="width:100%;background:var(--bg3);border:1px solid var(--border);border-radius:8px;padding:10px 14px;color:var(--text);font:inherit;resize:vertical;outline:none" placeholder="Подробное описание курса: что студент узнает, какие навыки получит, для кого этот курс..."></textarea></div>
        <div class="form-group">
          <label>Изображение обложки</label>
          <div style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">
            <input type="text" id="admin-img-url" name="imageUrl" placeholder="URL изображения..." style="flex:1">
            <label class="upload-btn">
              ${ic.camera} Загрузить файл
              <input type="file" accept="image/*" style="display:none" onchange="adminUploadCourseImage(this)">
            </label>
          </div>
        </div>
        <div style="display:flex;gap:10px">
          <button type="submit" class="btn btn-primary">${adminEditId?'Сохранить изменения':'Создать курс'}</button>
          <button type="button" class="btn btn-outline" onclick="adminEditId=null;showAdminForm=false;render()">Отмена</button>
        </div>
      </form>
    </div>` : `
    <div style="margin-bottom:20px">
      <button class="btn btn-primary" onclick="showAdminForm=true;adminEditId=null;render()">
        ${ic.plus} Добавить курс
      </button>
    </div>`}

    <div class="section-title" style="margin-bottom:14px">Все курсы (${adminCourses.length})</div>
    <div class="admin-course-list">
      ${adminCourses.map(c=>`
        <div class="card admin-course-row">
          <div class="admin-course-thumb" style="background:${courseThumbBg(c.id)}">
            ${c.imageUrl?`<img src="${esc(c.imageUrl)}" alt="" style="width:100%;height:100%;object-fit:cover;border-radius:6px">`:esc((c.title||'?')[0])}
          </div>
          <div class="admin-course-info">
            <div class="admin-course-title">${esc(c.title)}</div>
            <div style="font-size:12px;color:var(--text2)">${esc(c.category)} &nbsp;·&nbsp; ${esc(c.authorName)} &nbsp;·&nbsp; ${c.isPaid?c.price+' ₽':'Бесплатно'} &nbsp;·&nbsp; ${c.lessonsCount} уроков</div>
          </div>
          <div style="display:flex;gap:6px;flex-shrink:0">
            <button class="btn btn-sm btn-outline" onclick="adminEditCourse(${c.id})">${ic.edit} Изменить</button>
            <button class="btn btn-sm btn-danger" onclick="adminDeleteCourse(${c.id},'${esc(c.title)}')">${ic.trash}</button>
          </div>
        </div>`).join('')}
    </div>
  </div>`;
}

// ─── AUTH ACTIONS ────────────────────────────────────────────────────────────
async function doLogin(e) {
  e.preventDefault(); const f=e.target;
  const errEl=document.getElementById('auth-error'), btn=document.querySelector('#auth-btn');
  errEl.textContent=''; if(btn){btn.disabled=true;btn.textContent='Вход...';}
  try {
    const d=await api.post('/auth/login',{email:f.email.value,password:f.password.value});
    state.user=d.user; allCoursesData=null; coursesData=null; dashData=null; navigate('/');
  } catch(err) {
    errEl.textContent=err.message;
    if(btn){btn.disabled=false;btn.textContent='Войти';}
  }
}
async function doRegister(e) {
  e.preventDefault(); const f=e.target;
  const errEl=document.getElementById('auth-error');
  errEl.textContent='';
  try {
    const d=await api.post('/auth/register',{name:f.name.value,email:f.email.value,password:f.password.value});
    state.user=d.user; allCoursesData=null; coursesData=null; dashData=null; navigate('/');
  } catch(err) { errEl.textContent=err.message; }
}
async function logout() {
  await api.post('/auth/logout').catch(()=>{});
  state.user=null; dashData=null; allCoursesData=null; coursesData=null;
  courseDetail=null; profileProgress=null; profileEnrollments=null; adminCourses=null;
  navigate('/login');
}
async function saveProfile(e) {
  e.preventDefault(); const f=e.target;
  const errEl=document.getElementById('profile-error');
  errEl.textContent='';
  try {
    const avatarUrl = document.getElementById('avatar-url-val')?.value || null;
    const user=await api.patch('/users/profile',{
      name:f.name.value,
      dateOfBirth:f.dateOfBirth.value||null,
      avatarUrl: avatarUrl||null,
    });
    state.user=user; toast('Профиль обновлён'); render();
  } catch(err) { errEl.textContent=err.message; }
}

// ─── UTILS ───────────────────────────────────────────────────────────────────
function loading() {
  return `<div class="page"><div class="loading-page"><div class="spinner"></div></div></div>`;
}

// ─── RENDER ──────────────────────────────────────────────────────────────────
function render() {
  const app = document.getElementById('app');
  const { route, routeParam } = state;
  if (!state.user && route !== '/register') { app.innerHTML = renderLogin(); return; }
  if (route === '/login') { app.innerHTML = renderLogin(); return; }
  if (route === '/register') { app.innerHTML = renderRegister(); return; }

  let page;
  if (route==='/'||route==='') page = renderDashboard();
  else if (route==='/courses'&&!routeParam) page = renderCourses();
  else if (route==='/courses'&&routeParam) page = renderCourseDetail(routeParam);
  else if (route==='/profile') page = renderProfile();
  else if (route==='/admin') page = renderAdmin();
  else page = `<div class="page"><div class="loading-page" style="flex-direction:column;gap:8px"><div style="font-size:18px">404</div><div style="color:var(--text2)">Страница не найдена</div></div></div>`;

  app.innerHTML = renderSidebar() + `<div class="main-content">${page}</div>`;
}

// ─── BOOT ────────────────────────────────────────────────────────────────────
async function boot() {
  const { path, param } = getRoute();
  state.route = path; state.routeParam = param;
  document.getElementById('app').innerHTML = `<div style="display:flex;align-items:center;justify-content:center;height:100vh;width:100%"><div class="spinner"></div></div>`;
  try { state.user = await api.get('/auth/me'); } catch(e) { state.user = null; }
  render();
}
boot();
