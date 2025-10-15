document.addEventListener('DOMContentLoaded', () => {
    const authSection = document.getElementById('auth-section');
    const scheduleSection = document.getElementById('schedule-section');
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    const createScheduleForm = document.getElementById('create-schedule-form');
    const authInfo = document.getElementById('auth-info');

    const calendarContainer = document.getElementById('calendar-container');
    const calendarMonthYear = document.getElementById('calendar-month-year');
    const prevWeekBtn = document.getElementById('prev-week-btn');
    const weekHeader = document.getElementById('week-header');
    const timeAxis = document.getElementById('time-axis');
    const scheduleGrid = document.getElementById('schedule-grid');
    const nextWeekBtn = document.getElementById('next-week-btn');

    const showRegisterLink = document.getElementById('show-register-link');
    const showLoginLink = document.getElementById('show-login-link');
    const loginFormContainer = document.getElementById('login-form-container');
    const registerFormContainer = document.getElementById('register-form-container');

    const API_URL = 'http://localhost:8080/api';
    let currentUser = null;
    let currentDate = new Date();
    let schedulesCache = [];

    // --- Event Listeners ---
    prevWeekBtn.addEventListener('click', () => {
        currentDate.setDate(currentDate.getDate() - 7);
        renderWeeklyCalendar(schedulesCache);
    });

    nextWeekBtn.addEventListener('click', () => {
        currentDate.setDate(currentDate.getDate() + 7);
        renderWeeklyCalendar(schedulesCache);
    });

    // --- Utility Functions ---
    const getToken = () => localStorage.getItem('jwt_token');
    const setToken = (token) => localStorage.setItem('jwt_token', token);
    const clearToken = () => localStorage.removeItem('jwt_token');

    const parseJwt = (token) => {
        try {
            return JSON.parse(atob(token.split('.')[1]));
        } catch (e) {
            return null;
        }
    };

    // --- UI Toggling ---
    const showAuthUI = () => {
        authSection.classList.remove('hidden');
        scheduleSection.classList.add('hidden');
        authInfo.innerHTML = '';
    };

    const showScheduleUI = () => {
        authSection.classList.add('hidden');
        scheduleSection.classList.remove('hidden');
        const token = getToken();
        if (token) {
            currentUser = parseJwt(token);
            authInfo.innerHTML = `
                <p>ログイン中 (UserID: ${currentUser.user_id})</p>
                <button id="logout-btn">ログアウト</button>
            `;
            document.getElementById('logout-btn').addEventListener('click', handleLogout);
            fetchSchedules();

            // 管理者(user_id=1)の場合、管理者パネルを表示
            const adminPanel = document.getElementById('admin-panel');
            if (currentUser.user_id === 1) {
                adminPanel.classList.remove('hidden');
                fetchUsers();
            } else {
                adminPanel.classList.add('hidden');
            }
        }
    };

    showRegisterLink.addEventListener('click', (e) => {
        e.preventDefault();
        loginFormContainer.classList.add('hidden');
        registerFormContainer.classList.remove('hidden');
    });

    showLoginLink.addEventListener('click', (e) => {
        e.preventDefault();
        loginFormContainer.classList.remove('hidden');
        registerFormContainer.classList.add('hidden');
    });

    // --- API Call Functions ---
    const apiFetch = async (endpoint, options = {}) => {
        const token = getToken();
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_URL}${endpoint}`, { ...options, headers });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`API Error: ${response.status} ${errorText}`);
        }

        if (response.status === 204) { // No Content
            return null;
        }
        return response.json();
    };

    // --- Authentication Logic ---
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = document.getElementById('login-email').value;
        const password = document.getElementById('login-password').value;
        try {
            const data = await apiFetch('/users/login', {
                method: 'POST',
                body: JSON.stringify({ email, password }),
            });
            setToken(data.token);
            showScheduleUI();
        } catch (error) {
            alert(`Login failed: ${error.message}`);
        }
    });

    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('register-username').value;
        const email = document.getElementById('register-email').value;
        const password = document.getElementById('register-password').value;
        try {
            await apiFetch('/users/register', {
                method: 'POST',
                body: JSON.stringify({ username, email, password }),
            });
            alert('Registration successful! Please log in.');
            showLoginLink.click();
        } catch (error) {
            alert(`Registration failed: ${error.message}`);
        }
    });

    const handleLogout = () => {
        clearToken();
        currentUser = null;
        document.getElementById('admin-panel').classList.add('hidden');
        document.getElementById('admin-schedule-view').classList.add('hidden');
        showAuthUI();
    };

    // --- Schedule Logic ---
    const fetchSchedules = async () => {
        if (!currentUser) return;
        try {
            schedulesCache = await apiFetch(`/users/${currentUser.user_id}/schedules`) || [];
            renderWeeklyCalendar(schedulesCache);
        } catch (error) {
            alert(`Failed to fetch schedules: ${error.message}`);
            schedulesCache = [];
            renderWeeklyCalendar(schedulesCache);
        }
    };

    const renderWeeklyCalendar = (schedules) => {
        // Clear previous content
        weekHeader.innerHTML = '';
        timeAxis.innerHTML = '';
        scheduleGrid.innerHTML = '';

        const startOfWeek = new Date(currentDate);
        startOfWeek.setDate(startOfWeek.getDate() - (startOfWeek.getDay() || 7) + 1); // Monday

        const endOfWeek = new Date(startOfWeek);
        endOfWeek.setDate(endOfWeek.getDate() + 6); // Sunday

        calendarMonthYear.textContent = `${startOfWeek.getFullYear()}年 ${startOfWeek.getMonth() + 1}月`;

        // --- Render Time Axis ---
        for (let hour = 0; hour < 24; hour++) {
            const timeLabel = document.createElement('div');
            timeLabel.className = 'time-label';
            timeLabel.textContent = `${String(hour).padStart(2, '0')}:00`;
            timeAxis.appendChild(timeLabel);
        }

        const days = ['月', '火', '水', '木', '金', '土', '日'];
        const dayColumns = [];

        // --- Render Header and Day Columns ---
        // Add a blank corner for the time axis
        const corner = document.createElement('div');
        corner.className = 'header-corner';
        weekHeader.appendChild(corner);

        for (let i = 0; i < 7; i++) {
            const day = new Date(startOfWeek);
            day.setDate(day.getDate() + i);

            // Render Header
            const headerCell = document.createElement('div');
            headerCell.className = 'calendar-header-cell';
            headerCell.innerHTML = `
                <div class="day-of-week">${days[i]}</div>
                <div class="date-number">${day.getDate()}</div>
            `;
            if (isToday(day)) {
                headerCell.classList.add('today');
            }
            weekHeader.appendChild(headerCell);

            // Create Day Column
            const dayColumn = document.createElement('div');
            dayColumn.className = 'day-column';
            if (isToday(day)) {
                dayColumn.classList.add('today');
            }
            scheduleGrid.appendChild(dayColumn);
            dayColumns.push(dayColumn);
        }

        // --- Render Schedules ---
        schedules.forEach(s => {
            const startTime = new Date(s.start_time);
            const endTime = new Date(s.end_time);
            const dayIndex = (startTime.getDay() + 6) % 7; // Monday is 0

            // Ensure the schedule is within the current week
            if (startTime < startOfWeek || startTime > endOfWeek) {
                return;
            }

            const startMinutes = startTime.getHours() * 60 + startTime.getMinutes();
            const endMinutes = endTime.getHours() * 60 + endTime.getMinutes();
            const durationMinutes = Math.max(0, endMinutes - startMinutes);

            // Position and height calculation (1 minute = 1 pixel for simplicity)
            const top = startMinutes;
            const height = durationMinutes;

            const scheduleItem = document.createElement('div');
            scheduleItem.className = 'schedule-item-calendar';
            scheduleItem.style.top = `${top}px`;
            scheduleItem.style.height = `${height}px`;

            scheduleItem.innerHTML = `
                <div class="schedule-title">${s.title}</div>
                <div class="schedule-time">
                    ${startTime.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })} -
                    ${endTime.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })}
                </div>
                <button class="delete-btn-calendar" data-id="${s.id}">&times;</button>
            `;

            dayColumns[dayIndex].appendChild(scheduleItem);
        });

        setupDeleteButtons();
    };

    const isToday = (someDate) => {
        const today = new Date();
        return someDate.getDate() === today.getDate() &&
               someDate.getMonth() === today.getMonth() &&
               someDate.getFullYear() === today.getFullYear();
    };

    const setupDeleteButtons = () => {
        document.querySelectorAll('.delete-btn-calendar').forEach(button => {
            // Remove existing listeners to avoid duplicates
            button.replaceWith(button.cloneNode(true));
        });
        document.querySelectorAll('.delete-btn-calendar').forEach(button => {
            button.addEventListener('click', async (e) => {
                e.stopPropagation(); // Stop event bubbling
                const scheduleId = e.target.dataset.id;
                if (confirm('このスケジュールを本当に削除しますか？')) {
                    try {
                        await apiFetch(`/schedules/${scheduleId}`, { method: 'DELETE' });
                        fetchSchedules(); // Refresh calendar
                    } catch (error) {
                        alert(`Failed to delete schedule: ${error.message}`);
                    }
                }
            });
        });
    };

    createScheduleForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const title = document.getElementById('schedule-title').value;
        const startTime = document.getElementById('schedule-start-time').value;
        const endTime = document.getElementById('schedule-end-time').value;
        const description = document.getElementById('schedule-description').value;
        const location = document.getElementById('schedule-location').value;

        try {
            await apiFetch('/schedules', {
                method: 'POST',
                body: JSON.stringify({
                    title,
                    start_time: new Date(startTime).toISOString(),
                    end_time: new Date(endTime).toISOString(),
                    description,
                    location,
                    owner_id: currentUser.user_id,
                }),
            });
            createScheduleForm.reset();
            fetchSchedules();
        } catch (error) {
            alert(`Failed to create schedule: ${error.message}`);
        }
    });

    // --- Admin Logic ---
    const adminCreateScheduleForm = document.getElementById('admin-create-schedule-form');

    const fetchUsers = async () => {
        try {
            const users = await apiFetch('/admin/users');
            renderUsers(users);
        } catch (error) {
            alert(`Failed to fetch users: ${error.message}`);
        }
    };

    const renderUsers = (users) => {
        const userList = document.getElementById('user-list');
        userList.innerHTML = '';
        if (!users || users.length === 0) {
            userList.innerHTML = '<p>ユーザーが見つかりません。</p>';
            return;
        }
        users.forEach(user => {
            const item = document.createElement('div');
            item.className = 'user-item';
            item.innerHTML = `
                <span>${user.username} (ID: ${user.id}, Email: ${user.email})</span>
                <button class="manage-schedules-btn" data-user-id="${user.id}" data-user-name="${user.username}">
                    スケジュール管理
                </button>
            `;
            userList.appendChild(item);
        });

        document.querySelectorAll('.manage-schedules-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const userId = e.target.dataset.userId;
                const userName = e.target.dataset.userName;
                handleManageSchedules(userId, userName);
            });
        });
    };

    const handleManageSchedules = (userId, userName) => {
        const adminScheduleView = document.getElementById('admin-schedule-view');
        const selectedUserName = document.getElementById('admin-selected-user-name');
        const ownerIdInput = document.getElementById('admin-owner-id');

        selectedUserName.textContent = `${userName}のスケジュール`;
        ownerIdInput.value = userId;
        adminScheduleView.classList.remove('hidden');

        fetchAdminSchedules(userId);
    };

    const fetchAdminSchedules = async (userId) => {
        try {
            const schedules = await apiFetch(`/users/${userId}/schedules`);
            renderAdminSchedules(schedules, userId);
        } catch (error) {
            alert(`Failed to fetch schedules for user ${userId}: ${error.message}`);
            document.getElementById('admin-schedule-list').innerHTML = `<p class="error">スケジュールの読み込みに失敗しました。</p>`;
        }
    };

    const renderAdminSchedules = (schedules, ownerId) => {
        const adminScheduleList = document.getElementById('admin-schedule-list');
        adminScheduleList.innerHTML = '';
        if (!schedules || schedules.length === 0) {
            adminScheduleList.innerHTML = '<p>スケジュールはありません。</p>';
            return;
        }
        schedules.forEach(s => {
            const item = document.createElement('div');
            item.className = 'schedule-item';
            item.innerHTML = `
                <h4>${s.title}</h4>
                <p><strong>開始:</strong> ${new Date(s.start_time).toLocaleString()}</p>
                <p><strong>終了:</strong> ${new Date(s.end_time).toLocaleString()}</p>
                <p>${s.description || ''}</p>
                <p><em>場所: ${s.location || 'N/A'}</em></p>
                <div class="actions">
                    <button class="admin-delete-btn" data-id="${s.id}" data-owner-id="${ownerId}">削除</button>
                </div>
            `;
            adminScheduleList.appendChild(item);
        });

        document.querySelectorAll('.admin-delete-btn').forEach(button => {
            button.addEventListener('click', async (e) => {
                const scheduleId = e.target.dataset.id;
                const currentOwnerId = e.target.dataset.ownerId;
                if (confirm('このスケジュールを本当に削除しますか？')) {
                    try {
                        await apiFetch(`/schedules/${scheduleId}`, { method: 'DELETE' });
                        fetchAdminSchedules(currentOwnerId);
                    } catch (error) {
                        alert(`Failed to delete schedule: ${error.message}`);
                    }
                }
            });
        });
    };

    adminCreateScheduleForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const ownerId = document.getElementById('admin-owner-id').value;
        const title = document.getElementById('admin-schedule-title').value;
        const startTime = document.getElementById('admin-schedule-start-time').value;
        const endTime = document.getElementById('admin-schedule-end-time').value;
        const description = document.getElementById('admin-schedule-description').value;
        const location = document.getElementById('admin-schedule-location').value;

        if (!ownerId) {
            alert('ユーザーが選択されていません。');
            return;
        }

        try {
            await apiFetch('/schedules', {
                method: 'POST',
                body: JSON.stringify({
                    title,
                    start_time: new Date(startTime).toISOString(),
                    end_time: new Date(endTime).toISOString(),
                    description,
                    location,
                    owner_id: parseInt(ownerId, 10),
                }),
            });
            adminCreateScheduleForm.reset();
            // keep ownerId hidden input value
            document.getElementById('admin-owner-id').value = ownerId;
            fetchAdminSchedules(ownerId);
        } catch (error) {
            alert(`Failed to create schedule: ${error.message}`);
        }
    });

    // --- Initial Check ---
    if (getToken()) {
        showScheduleUI();
    } else {
        showAuthUI();
    }
});