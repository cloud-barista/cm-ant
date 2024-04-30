
document.addEventListener('DOMContentLoaded', function () {

    let loadTestKey = null;


    const rows = document.querySelectorAll('table tr:not(:first-child)');
    rows.forEach(row => {
        row.addEventListener('click', function () {
            loadTestKey = this.cells[2].textContent;
            
            Promise.all([
                fetchAggregate(),
                fetchListData()
            ])
            .then(results => {
                console.log('fetch data clear');
            })
            .catch(error => {
                console.error('Error fetching list data:', error);
            });
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
    
    function fetchListData() {
        return fetch(`/ant/api/v1/load/result?loadTestKey=${loadTestKey}`)
            .then(response => response.json())
            .then(data => {
                drawChart(data);
                return null; // 데이터 반환
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
                    <td>${data.average.toFixed(2)}</td>
                    <td>${data.median.toFixed(2)}</td>
                    <td>${data.ninetyPercent.toFixed(2)}</td>
                    <td>${data.ninetyFive.toFixed(2)}</td>
                    <td>${data.ninetyNine.toFixed(2)}</td>
                    <td>${data.minTime.toFixed(2)}</td>
                    <td>${data.maxTime.toFixed(2)}</td>
                    <td>${data.errorPercent.toFixed(2)}%</td>
                    <td>${data.throughput.toFixed(2)}</td>
                    <td>${data.receivedKB.toFixed(2)}</td>
                    <td>${data.sentKB.toFixed(2)}</td>
                </tr>
            `;
            tbody.innerHTML += row;  // 생성된 행을 tbody에 추가
        });
        staticDiv.style.display = 'block';
    }

    let myChart = [];
    

    function drawChart(data) {
        const staticDiv = document.getElementById('chartDiv');
        staticDiv.style.display = 'none';
        
        if (!myChart) {
            myChart.map(s => s.destroy());
            myChart = [];
        }
        staticDiv.innerHTML = '';
        
        // const timestamps = Object.keys(data).map(ts => new Date(ts).toLocaleTimeString());
        const labels = Object.keys(data).flat();

        labels.map(label => {

            const canvas = document.createElement('canvas');
            canvas.id = 'canvas-' + label;

            const ctx = canvas.getContext('2d');
            const v = data[label];
            const count = Math.max(60, v.length / 1000)
            const countedData = filterData(v, count);

            const elapsedTimes = countedData.map(item => item.Elapsed);
            const bytes = countedData.map(item => item.Bytes);
            const sendBytes = countedData.map(item => item.SentBytes);
            const latencies = countedData.map(item => item.Latency);
            const idleTimes = countedData.map(item => item.IdleTime);
            const connections = countedData.map(item => item.Connection);
            const timestamps = countedData.map(item => new Date(item.Timestamp).toLocaleTimeString());


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
                                maxTicksLimit: count
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

            myChart.push(charttt) ;
            staticDiv.appendChild(canvas);
        })

        

        staticDiv.style.display = 'block';
    }

    function filterData(data, count) {

        if (data.length < count) {
            return data;
        }
        const interval = Math.floor(data.length / count);
        let datas = [];
        for (let i = 0; i < data.length; i += interval) {
            datas.push(data[i]);
        }
        return datas;
    }

    const button = document.getElementById('refresh-chart');

    button.addEventListener('click', function() {

        if (!loadTestKey) {
            return;
        }
        Promise.all([
            fetchAggregate(),
            fetchListData()
        ])
        .then(results => {
            console.log('fetch data clear');
        })
        .catch(error => {
            console.error('Error fetching list data:', error);
        });
    });

});

