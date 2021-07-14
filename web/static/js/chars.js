$(document).ready(function () {

    var w=$(window).outerWidth();
    var isMobile = {Android: function() {return navigator.userAgent.match(/Android/i);},BlackBerry: function() {return navigator.userAgent.match(/BlackBerry/i);},iOS: function() {return navigator.userAgent.match(/iPhone|iPad|iPod/i);},Opera: function() {return navigator.userAgent.match(/Opera Mini/i);},Windows: function() {return navigator.userAgent.match(/IEMobile/i);},any: function() {return (isMobile.Android() || isMobile.BlackBerry() || isMobile.iOS() || isMobile.Opera() || isMobile.Windows());}};

    Highcharts.getJSON('./get-credit-char.json', function (data) {
        Highcharts.chart('credit-container', {

            chart: {
                type: 'column'
            },

            title: {
                text: ''
            },

            xAxis: {
                categories: data.categories,
            },

            yAxis: {
                allowDecimals: false,
                min: 0,
                title: {
                    text: 'Деньги'
                }
            },

            legend: {
                enabled: !isMobile || w > 1024,
            },

            tooltip: {
                headerFormat: '<b>{point.x}</b><br/>',
                formatter: function () {
                    return '<b>' + this.x + '</b><br/>' +
                        this.series.name + ': ' + this.y + ' ₽<br/>' +
                        'Всего за месяц: ' + this.point.stackTotal +' ₽';
                }
            },

            plotOptions: {
                column: {
                    stacking: 'normal',
                    // dataLabels: {
                    //     enabled: true
                    // },
                } ,
            },

            series: data.series,
        })
    });

    Highcharts.getJSON('./get-debit-char.json', function (data) {
        Highcharts.chart('debit-container', {

            chart: {
                type: 'column'
            },

            title: {
                text: ''
            },

            xAxis: {
                categories: data.categories,
            },

            yAxis: {
                allowDecimals: false,
                min: 0,
                title: {
                    text: 'Деньги'
                }
            },

            legend: {
                enabled: !isMobile || w > 1024,
            },

            tooltip: {
                formatter: function () {
                    return '<b>' + this.x + '</b><br/>' +
                        this.series.name + ': ' + this.y + ' ₽<br/>' +
                        'Всего за месяц: ' + this.point.stackTotal +' ₽';
                }
            },

            plotOptions: {
                column: {
                    stacking: 'normal',
                    // dataLabels: {
                    //     enabled: true
                    // },
                }
            },

            series: data.series,
        })
    });

    Highcharts.getJSON('./get-debit-credit-line-char.json', function (data) {
        Highcharts.stockChart('debit-credit-line-char-container', {
            rangeSelector: {
                selected: 1
            },

            title: {
                text: ''
            },

            plotOptions: {
                line: {
                    dataLabels: {
                        enabled: !isMobile || w > 1024
                    },

                }
            },

            legend: {
                enabled: !isMobile || w > 1024,
            },

            tooltip: {
                valueSuffix: ' ₽',
            },

            series: [
                {
                    name: 'Приход',
                    data: data[0],
                    tooltip: {
                        valueDecimals: 0
                    },
                    // color: '#2f0',

                },
                {
                    name: 'Расход',
                    data: data[1],
                    tooltip: {
                        valueDecimals: 0
                    },
                    // color: '#fc4040',
                },
            ]
        });
    });

});