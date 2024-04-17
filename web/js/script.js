
const localButton = document.getElementById('local-button');
const remoteButton = document.getElementById('remote-button');
const remoteConfigForm = document.getElementById('remote-config-form');

localButton.addEventListener('click', function () {
    remoteConfigForm.style.display = 'none';
    localButton.classList.add('active');
    remoteButton.classList.remove('active');
    // Additional logic for switching to local configuration can go here
});

remoteButton.addEventListener('click', function () {
    remoteConfigForm.style.display = 'block';
    remoteButton.classList.add('active');
    localButton.classList.remove('active');
    // Additional logic for switching to remote configuration can go here
});

document.getElementById('remote-connection-type').addEventListener('change', function () {
    // Hide all input containers first
    document.getElementById('auth-container').style.display = 'none';
    document.getElementById('built-in-container').style.display = 'none';

    // Determine which set of inputs to show based on the selected connection type
    switch (this.value) {
        case 'password':
        case 'privateKey':
            // Show the container for Password/Private Key inputs
            document.getElementById('auth-container').style.display = 'block';
            break;
        case 'builtIn':
            // Show the container for Built-In inputs
            document.getElementById('built-in-container').style.display = 'block';
            break;
    }
});

function getInstallLocation() {
    var localButton = document.getElementById('local-button');
    return localButton.classList.contains('active') ? 'local' : 'remote';
}

// ---------------------------------------

const updateSliderValue = (sliderId, displayId) => {
    const slider = document.getElementById(sliderId);
    const display = document.getElementById(displayId);
    display.textContent = slider.value;

    slider.addEventListener('input', () => {
        display.textContent = slider.value;
    });
};

// Initialize slider values and add event listeners
updateSliderValue('virtual-users', 'virtual-users-value');
updateSliderValue('duration', 'duration-value');
updateSliderValue('ramp-up-time', 'ramp-up-time-value');
updateSliderValue('ramp-up-steps', 'ramp-up-steps-value');


// ------------------------------------------------------------

const ctx = document.getElementById('load-chart').getContext('2d');
const loadChart = new Chart(ctx, {
    type: 'line',
    data: {
        datasets: [{
            label: 'Virtual Users',
            borderColor: 'rgb(54, 162, 235)',
            backgroundColor: 'rgba(54, 162, 235, 0.2)',
            fill: true
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: true,
        scales: {
            x: {
                type: 'linear',
                position: 'bottom',
                title: {
                    display: true,
                    text: 'Duration (s)'
                }
            },
            y: {
                beginAtZero: true,
                title: {
                    display: true,
                    text: 'Virtual Users'
                }
            }
        }
    }
});


function createGradient(ctx, area, rampUpTime, duration) {
    const gradient = ctx.createLinearGradient(area.left, 0, area.right, 0);
    const rampUpRatio = (rampUpTime / duration);

    // Add three color stops, with a clear separation at the ramp-up point
    gradient.addColorStop(0, 'rgba(153, 102, 255, 0.2)'); // Ramp-up color
    gradient.addColorStop(rampUpRatio, 'rgba(153, 102, 255, 0.2)'); // Ramp-up color
    gradient.addColorStop(rampUpRatio, 'rgba(75, 192, 192, 0.2)'); // Post ramp-up color
    gradient.addColorStop(1, 'rgba(75, 192, 192, 0.2)'); // Post ramp-up color

    return gradient;
}

function updateChart() {
    // Get the values from the input sliders
    const virtualUsers = parseInt(document.getElementById('virtual-users').value, 10);
    const duration = parseInt(document.getElementById('duration').value, 10);
    const rampUpTime = parseInt(document.getElementById('ramp-up-time').value, 10);
    const rampUpSteps = parseInt(document.getElementById('ramp-up-steps').value, 10);

    const stepDuration = rampUpTime / rampUpSteps;
    const usersPerStep = virtualUsers / rampUpSteps;

    // Data for the chart
    let chartData = [];

    chartData.push({
        x: stepDuration * 0,
        y: usersPerStep * 0
    });

    for (let i = 0; i < rampUpSteps; i++) {
        chartData.push({
            x: stepDuration * i,
            y: usersPerStep * (i + 1)
        });

        if ((i + 1) !== rampUpSteps) {
            chartData.push({
                x: stepDuration * (i + 1),
                y: usersPerStep * (i + 1)
            });
        }
    }

    chartData.push({
        x: rampUpTime,
        y: virtualUsers
    });

    chartData.push({
        x: duration + rampUpTime,
        y: virtualUsers
    });


    // Update chart data
    loadChart.data.labels = chartData.map(data => `${data.x}s`);
    loadChart.data.datasets[0].data = chartData;

    // Use the chart's context and the scale dimensions to create the gradient
    const gradient = createGradient(loadChart.ctx, loadChart.chartArea, rampUpTime, duration + rampUpTime);

    // Update the chart dataset to use the gradient for the backgroundColor property
    loadChart.data.datasets[0].backgroundColor = gradient;

    // Finally, update the chart to redraw with the new gradient
    loadChart.update();
}

document.getElementById('virtual-users').addEventListener('input', updateChart);
document.getElementById('duration').addEventListener('input', updateChart);
document.getElementById('ramp-up-time').addEventListener('input', updateChart);
document.getElementById('ramp-up-steps').addEventListener('input', updateChart);

// Initialize the chart with default slider values
updateChart();




// -------------------------------------------------------
const addButton = document.getElementById('add-load-test-detail');
const removeButton = document.getElementById('remove-load-test-detail');
const container = document.querySelector('.load-execution-http-req-container');

addButton.addEventListener('click', function () {
    // Clone the template
    const template = document.getElementById('load-execution-http-req-template');
    const clone = template.cloneNode(true);

    // Remove the id from the cloned element to avoid duplicate ids
    clone.removeAttribute('id');

    // Append the cloned template to the container
    container.appendChild(clone);
});

removeButton.addEventListener('click', function () {
    const allReqs = container.querySelectorAll('.load-execution-http-req');
    // Allow removal if there's more than one form section
    if (allReqs.length > 1) {
        container.removeChild(allReqs[allReqs.length - 1]);
    }
});


document.getElementById('execute-test').addEventListener('click', function () {
    var testName = document.getElementById('test-name-input').value;

    var installLocation = getInstallLocation();

    var remoteConnectionType = document.getElementById('remote-connection-type').value;
    var username = document.getElementById('remote-config-form').style.display !== 'none'
        ? document.getElementById('username').value
        : document.getElementById('username-built-in').value;
    var publicId = document.getElementById('public-id').value;
    var cert = document.getElementById('cert').value;
    var nsId = document.getElementById('ns-id').value;
    var mcisId = document.getElementById('mcis-id').value;

    var virtualUsers = document.getElementById('virtual-users').value;
    var duration = document.getElementById('duration').value;
    var rampUpTime = document.getElementById('ramp-up-time').value;
    var rampUpSteps = document.getElementById('ramp-up-steps').value;


    var httpReqs = Array.from(document.querySelectorAll('.load-execution-http-req')).map(req => {
        return {
            method: req.querySelector('[name="method"]').value,
            protocol: req.querySelector('[name="protocol"]').value,
            hostname: req.querySelector('[name="hostname"]').value,
            port: req.querySelector('[name="port"]').value,
            path: req.querySelector('[name="path"]').value,
            bodyData: req.querySelector('[name="bodyData"]').value
        };
    });


    var loadExecutionConfigReq = {
        testName: testName,
        virtualUsers: virtualUsers,
        duration: duration,
        rampUpTime: rampUpTime,
        rampUpSteps: rampUpSteps,
        httpReqs: httpReqs,
        loadEnvReq: {
            installLocation: installLocation,
            remoteConnectionType: remoteConnectionType,
            username: username,
            publicIp: publicId,
            cert: cert,
            nsId: nsId,
            mcisId: mcisId
        }
    };



    fetch('/ant/load/start', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(loadExecutionConfigReq)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        console.log('Load test execution started:', data);
    })
    .catch(error => {
        console.error('Error starting load test:', error);
    });
});



