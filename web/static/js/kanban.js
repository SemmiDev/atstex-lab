document.addEventListener('DOMContentLoaded', () => {
    const board = document.getElementById('kanban-board');
    const modal = document.getElementById('job-modal');
    const addJobBtn = document.getElementById('add-job-btn');
    const cancelBtn = document.getElementById('cancel-btn');
    const saveJobBtn = document.getElementById('save-job-btn');
    const jobForm = document.getElementById('job-form');
    const cvProfileSelect = document.getElementById('cv-profile');

    let applications = [];
    let cvProfiles = [];

    // --- Modal Handling ---
    const showModal = () => modal.classList.remove('hidden');
    const hideModal = () => modal.classList.add('hidden');

    addJobBtn.addEventListener('click', () => {
        jobForm.reset();
        document.getElementById('application-id').value = '';
        document.getElementById('deadline').value = '';
        document.getElementById('modal-title').textContent = 'Tambah Lamaran Pekerjaan';
        showModal();
    });

    cancelBtn.addEventListener('click', hideModal);
    document.getElementById('cancel-btn-top')?.addEventListener('click', hideModal);

    // --- Stats Pipeline ---
    const updateStats = () => {
        const counts = { Applied: 0, Interview: 0, Offer: 0, Rejected: 0 };
        applications.forEach(app => {
            if (counts.hasOwnProperty(app.status)) counts[app.status]++;
        });
        document.getElementById('stats-applied').textContent = counts.Applied;
        document.getElementById('stats-interview').textContent = counts.Interview;
        document.getElementById('stats-offer').textContent = counts.Offer;
        document.getElementById('stats-rejected').textContent = counts.Rejected;
    };

    // --- Helper: check if a date string is overdue ---
    const isOverdue = (deadline) => {
        if (!deadline) return false;
        const deadlineDate = new Date(deadline);
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        deadlineDate.setHours(0, 0, 0, 0);
        return deadlineDate < today;
    };

    const formatDeadline = (deadline) => {
        if (!deadline) return '';
        const d = new Date(deadline);
        return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
    };

    // --- Data Fetching ---
    const fetchCVProfiles = async () => {
        try {
            const response = await fetch('/api/cv-profiles');
            if (!response.ok) throw new Error('Failed to fetch CV profiles');
            cvProfiles = await response.json();
            populateCVProfileOptions();
        } catch (error) {
            console.error('Error fetching CV profiles:', error);
        }
    };

    const fetchApplications = async () => {
        try {
            const response = await fetch('/api/applications');
            if (!response.ok) throw new Error('Failed to fetch applications');
            applications = await response.json();
            renderBoard();
            updateStats();
        } catch (error) {
            console.error('Error fetching applications:', error);
        }
    };

    // --- Rendering ---
    const populateCVProfileOptions = () => {
        cvProfileSelect.innerHTML = '<option value="">Select a profile</option>';
        cvProfiles.forEach(profile => {
            const option = document.createElement('option');
            option.value = profile.id;
            option.textContent = profile.title;
            cvProfileSelect.appendChild(option);
        });
    };

    const createApplicationCard = (app) => {
        const card = document.createElement('div');
        card.className = 'kanban-card';
        if (isOverdue(app.deadline) && app.status !== 'Rejected' && app.status !== 'Offer') {
            card.classList.add('overdue');
        }
        card.dataset.id = app.id;

        let deadlineHtml = '';
        if (app.deadline) {
            const overdueClass = isOverdue(app.deadline) && app.status !== 'Rejected' && app.status !== 'Offer' ? 'deadline-overdue' : '';
            deadlineHtml = `<span class="deadline-label ${overdueClass}"><i class="ph ph-calendar"></i> ${formatDeadline(app.deadline)}</span>`;
        }

        card.innerHTML = `
            <h3>${app.company}</h3>
            <p>${app.job_title}</p>
            ${deadlineHtml}
            <button class="delete-app-btn" title="Delete"><i class="ph-bold ph-trash"></i></button>
        `;

        card.addEventListener('click', () => {
            document.getElementById('application-id').value = app.id;
            document.getElementById('company').value = app.company;
            document.getElementById('job-title').value = app.job_title;
            document.getElementById('cv-profile').value = app.cv_profile_id || '';
            document.getElementById('notes').value = app.notes || '';
            // Set deadline value (extract date part from ISO string)
            if (app.deadline) {
                const d = new Date(app.deadline);
                document.getElementById('deadline').value = d.toISOString().split('T')[0];
            } else {
                document.getElementById('deadline').value = '';
            }
            document.getElementById('modal-title').textContent = 'Edit Job Application';
            showModal();
        });

        card.querySelector('.delete-app-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            deleteApplication(app.id);
        });
        return card;
    };

    const renderBoard = () => {
        // Clear existing cards
        board.querySelectorAll('.card-container').forEach(container => container.innerHTML = '');

        // Render cards into columns
        applications.forEach(app => {
            const column = board.querySelector(`.kanban-column[data-status="${app.status}"]`);
            if (column) {
                const cardContainer = column.querySelector('.card-container');
                cardContainer.appendChild(createApplicationCard(app));
            }
        });
    };

    // --- API Calls ---
    saveJobBtn.addEventListener('click', async () => {
        const company = document.getElementById('company').value;
        const jobTitle = document.getElementById('job-title').value;
        const cvProfileId = document.getElementById('cv-profile').value;
        const notes = document.getElementById('notes').value;
        const deadline = document.getElementById('deadline').value || null;
        const appId = document.getElementById('application-id').value;

        if (!company || !jobTitle || !cvProfileId) {
            alert('Please fill in all required fields.');
            return;
        }

        const applicationData = {
            company,
            job_title: jobTitle,
            cv_profile_id: cvProfileId,
            notes,
            deadline,
        };

        try {
            if (appId) {
                const response = await fetch(`/api/applications/${appId}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(applicationData),
                });
                if (!response.ok) throw new Error('Failed to update application');

                const index = applications.findIndex(a => a.id === appId);
                if (index !== -1) {
                    applications[index] = { ...applications[index], ...applicationData };
                }
            } else {
                const response = await fetch('/api/applications', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(applicationData),
                });
                if (!response.ok) throw new Error('Failed to save application');
                const newApp = await response.json();
                applications.push(newApp);
            }
            renderBoard();
            updateStats();
            hideModal();
        } catch (error) {
            console.error('Error saving application:', error);
            alert('Could not save application.');
        }
    });

    const updateApplicationStatus = async (appId, newStatus) => {
        try {
            const response = await fetch(`/api/applications/${appId}/status`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status: newStatus }),
            });
            if (!response.ok) throw new Error('Failed to update status');
            // Update local data and re-render
            const app = applications.find(a => a.id === appId);
            if(app) app.status = newStatus;
            renderBoard();
            updateStats();
        } catch (error) {
            console.error('Error updating status:', error);
        }
    };

    const deleteApplication = async (appId) => {
        if (!confirm('Are you sure you want to delete this application?')) return;

        try {
            const response = await fetch(`/api/applications/${appId}`, { method: 'DELETE' });
            if (!response.ok) throw new Error('Failed to delete application');

            applications = applications.filter(a => a.id !== appId);
            renderBoard();
            updateStats();
        } catch (error) {
            console.error('Error deleting application:', error);
            alert('Could not delete application.');
        }
    };


    // --- Drag and Drop ---
    const initSortable = () => {
        const columns = document.querySelectorAll('.card-container');
        columns.forEach(container => {
            new Sortable(container, {
                group: 'kanban',
                animation: 150,
                ghostClass: 'sortable-ghost',
                dragClass: 'sortable-drag',
                onEnd: (evt) => {
                    const itemEl = evt.item;
                    const newStatus = evt.to.closest('.kanban-column').dataset.status;
                    const appId = itemEl.dataset.id;

                    // Update the status on the backend
                    updateApplicationStatus(appId, newStatus);
                },
            });
        });
    };

    // --- Initialization ---
    const init = async () => {
        await fetchCVProfiles();
        await fetchApplications();
        initSortable();
    };

    init();
});
