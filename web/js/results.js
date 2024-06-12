
document.addEventListener('DOMContentLoaded', function () {

    let loadTestKey = null;
    let loading = false;

    const rows = document.querySelectorAll('table tr:not(:first-child)');

    function fetchResult(cellValue) {
        if (loading === true) {
            alert("Load test result is fetching!");
            return;
        }

        loading = true;

        if (cellValue !== undefined) {
            loadTestKey = cellValue;
        }

        Promise.all([
            fetchAggregate(),
            fetchResultList(),
            fetchMetrics(),
        ])
            .then(results => {
                loading = false;
                console.log('fetch data clear');
            })
            .catch(error => {
                loading = false;
                console.error('Error fetching list data:', error);
            });
    }

    rows.forEach(row => {
        row.addEventListener('click', function () {
            const cellValue = this.cells[2].textContent;
            fetchResult(cellValue);
        });
    })

    function fetchAggregate() {
        return fetch(`/ant/api/v1/load/result?loadTestKey=${loadTestKey}&format=aggregate`)
            .then(response => response.json())
            .then(data => {
                updateTable(data);
                return null;
            })
            .catch(error => {
                throw error; // 오류를 다시 throw하여 Promise.all에서 catch로 넘김
            });
    }
    
    function fetchResultList() {
        return fetch(`/ant/api/v1/load/result?loadTestKey=${loadTestKey}`)
            .then(response => response.json())
            .then(data => {
                drawResultChart(data);
                return null; // 데이터 반환
            })
            .catch(error => {
                
                throw error; // 오류를 다시 throw하여 Promise.all에서 catch로 넘김
            });
    }

    function fetchMetrics() {
        return fetch(`/ant/api/v1/load/result/metrics?loadTestKey=${loadTestKey}`)
            .then(response => response.json())
            .then(data => {
                drawMetricsChart(data);
                return null;
            })
            .catch(error => {
                throw error; // 오류를 다시 throw하여 Promise.all에서 catch로 넘김
            });
    }

    function updateTable(data) {
        const staticDiv = document.getElementById('statisticsDiv');
        staticDiv.style.display = 'none';
        const tbody = document.getElementById('statsBody');
        tbody.innerHTML = '';

        data.forEach(data => {
            const row = `
                <tr>
                    <td>${data.label}</td>
                    <td>${data.requestCount}</td>
                    <td>${data.average.toFixed(2)} ms</td>
                    <td>${data.median.toFixed(2)} ms</td>
                    <td>${data.ninetyPercent.toFixed(2)} ms</td>
                    <td>${data.ninetyFive.toFixed(2)} ms</td>
                    <td>${data.ninetyNine.toFixed(2)} ms</td>
                    <td>${data.minTime.toFixed(2)} ms</td>
                    <td>${data.maxTime.toFixed(2)} ms</td>
                    <td>${data.errorPercent.toFixed(2)} %</td>
                    <td>${data.throughput.toFixed(2)} /sec</td>
                    <td>${data.receivedKB.toFixed(2)} /sec</td>
                    <td>${data.sentKB.toFixed(2)} /sec</td>
                </tr>
            `;
            tbody.innerHTML += row;  // 생성된 행을 tbody에 추가
        });
        staticDiv.style.display = 'block';
    }

    let metricsCharts = [];

    function drawMetricsChart(data) {
        const metricsDiv = document.getElementById('metrics-chart-div');
        metricsDiv.style.display = 'none';
        
        if (!metricsCharts) {
            metricsCharts.map(s => s.destroy());
            metricsCharts = [];
        }
        metricsDiv.innerHTML = '';

        const labels = Object.keys(data).flat();

        const sampleRate = 50;

        labels.map(label => {
            const canvas = document.createElement('canvas');

            canvas.style.width = '25%';
            canvas.style.minWidth = '200px';

            canvas.id = 'canvas-' + label;
            const ctx = canvas.getContext('2d');
            const v = data[label];
            const sampledData = sampleData(v, sampleRate);

            const values = sampledData.map(item => item.Value);
            const timestamps = sampledData.map(item => new Date(item.Timestamp).toLocaleTimeString());

            const charttt = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: timestamps, 
                    datasets: [
                        {
                            label: label + ' (' + sampledData[0].Unit + ')',
                            data: values,
                            borderColor: 'rgb(255, 87, 34)',
                            backgroundColor: 'rgba(255, 87, 34, 0.5)',
                            fill: false
                        },
                    ]
                },
                options: {
                    scales: {
                        x: {
                            ticks: {
                                maxTicksLimit: sampleRate
                            }
                        },
                        y: {
                            beginAtZero: true
                        }
                    },
                    responsive: true,
                    maintainAspectRatio: true
                }
            });

            metricsCharts.push(charttt) ;
            metricsDiv.appendChild(canvas);
        })

        

        metricsDiv.style.display = 'block';
    }

    let resultCharts = [];

    function drawResultChart(data) {
        const staticDiv = document.getElementById('result-chart-div');
        staticDiv.style.display = 'none';
        
        if (!resultCharts) {
            resultCharts.map(s => s.destroy());
            resultCharts = [];
        }
        staticDiv.innerHTML = '';

        // const timestamps = Object.keys(data).map(ts => new Date(ts).toLocaleTimeString());
        const labels = Object.keys(data).flat();

        const sampleRate = 100;

        labels.map(label => {
            const canvas = document.createElement('canvas');

            canvas.id = 'canvas-' + label;
            const ctx = canvas.getContext('2d');
            const v = data[label];
            const sampledData = sampleData(v, sampleRate);

            const elapsedTimes = sampledData.map(item => item.Elapsed);
            const bytes = sampledData.map(item => item.Bytes);
            const sendBytes = sampledData.map(item => item.SentBytes);
            const latencies = sampledData.map(item => item.Latency);
            const idleTimes = sampledData.map(item => item.IdleTime);
            const connections = sampledData.map(item => item.Connection);
            const timestamps = sampledData.map(item => new Date(item.Timestamp).toLocaleTimeString());


            const charttt = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: timestamps, 
                    datasets: [
                        {
                            label: 'Elapsed Time (ms)',
                            data: elapsedTimes,
                            borderColor: 'rgb(255, 87, 34)',
                            backgroundColor: 'rgba(255, 87, 34, 0.5)',
                            fill: false
                        },
                        {
                            label: 'Bytes (kb)',
                            data: bytes,
                            borderColor: 'rgb(33, 150, 243)',
                            backgroundColor: 'rgba(33, 150, 243, 0.5)',
                            fill: false
                        },
                        {
                            label: 'Send Bytes (kb)',
                            data: sendBytes,
                            borderColor: 'rgb(139, 195, 74)',
                            backgroundColor: 'rgba(139, 195, 74, 0.5)',
                            fill: false
                        },
                        {
                            label: 'Idle Time (ms)',
                            data: idleTimes,
                            borderColor: 'rgb(156, 39, 176)',
                            backgroundColor: 'rgba(156, 39, 176, 0.5)',
                            fill: false
                        },
                        {
                            label: 'Latency (ms)',
                            data: latencies,
                            borderColor: 'rgb(255, 193, 7)',
                            backgroundColor: 'rgba(255, 193, 7, 0.5)',
                            fill: false
                        },
                        {
                            label: 'Connection Time (ms)',
                            data: connections,
                            borderColor: 'rgb(96, 125, 139)',
                            backgroundColor: 'rgba(96, 125, 139, 0.5)',
                            fill: false
                        }
                    ]
                },
                options: {
                    plugins: {
                        zoom: {
                            zoom: {
                                wheel: {
                                    enabled: true, 
                                },
                                pinch: {
                                    enabled: true
                                },
                                mode: 'x'
                            }
                        }
                    },
                    scales: {
                        x: {
                            ticks: {
                                maxTicksLimit: sampleRate
                            }
                        },
                        y: {
                            beginAtZero: true
                        }
                    },
                    responsive: true,
                    maintainAspectRatio: true
                }
            });

            resultCharts.push(charttt) ;
            staticDiv.appendChild(canvas);
        })

        

        staticDiv.style.display = 'block';
    }


    function sampleData(data, sampleRate) {
        if (data.length < sampleRate) {
            return data;
        }
        const interval = Math.floor(data.length / sampleRate);
        let sampledData = [];
        for (let i = 0; i < data.length; i += interval) {
            sampledData.push(data[i]);
        }
        return sampledData;
    }

    const button = document.getElementById('refresh-chart');

    button.addEventListener('click', function() {
        fetchResult();
    });


    const stopButton = document.getElementById('stop-load-test');

    stopButton.addEventListener('click', function(event) {
        stopLoadTest();
        event.stopPropagation();
    });

    function stopLoadTest() {

        if (!loadTestKey) {
            return
        }

        fetch(`/ant/api/v1/load/stop`,{
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({"loadTestKey": loadTestKey})
        })
            .then(response => response.json())
            .then(data => {
                console.log(data);
                console.log('load test stopped!');
                return null;
            })
            .catch(error => {
                throw error;
            });

    }


});

