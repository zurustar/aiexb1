document.addEventListener('DOMContentLoaded', () => {
    const authSection = document.getElementById('auth-section');
    const scheduleSection = document.getElementById('schedule-section');
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    const createScheduleForm = document.getElementById('create-schedule-form');
    const scheduleList = document.getElementById('schedule-list');
    const authInfo = document.getElementById('auth-info');

    const showRegisterLink = document.getElementById('show-register-link');
    const showLoginLink = document.getElementById('show-login-link');
    const loginFormContainer = document.getElementById('login-form-container');
    const registerFormContainer = document.getElementById('register-form-container');

    const API_URL = 'http://localhost:8080/api';
    let currentUser = null;

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
        showAuthUI();
    };

    // --- Schedule Logic ---
    const fetchSchedules = async () => {
        if (!currentUser) return;
        try {
            const schedules = await apiFetch(`/users/${currentUser.user_id}/schedules`);
            renderSchedules(schedules);
        } catch (error) {
            alert(`Failed to fetch schedules: ${error.message}`);
        }
    };

    const renderSchedules = (schedules) => {
        scheduleList.innerHTML = '';
        if (!schedules || schedules.length === 0) {
            scheduleList.innerHTML = '<p>スケジュールはありません。</p>';
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
                    <button class="delete-btn" data-id="${s.id}">削除</button>
                </div>
            `;
            scheduleList.appendChild(item);
        });

        // Add event listeners for delete buttons
        document.querySelectorAll('.delete-btn').forEach(button => {
            button.addEventListener('click', async (e) => {
                const scheduleId = e.target.dataset.id;
                if (confirm('このスケジュールを本当に削除しますか？')) {
                    try {
                        await apiFetch(`/schedules/${scheduleId}`, { method: 'DELETE' });
                        fetchSchedules(); // Refresh list
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


    // --- Initial Check ---
    if (getToken()) {
        showScheduleUI();
    } else {
        showAuthUI();
    }
});