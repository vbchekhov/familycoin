$(document).ready(function () {

    $(".navbar-burger").click(function () {
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    });

    $(".open-reciept").click(function () {
        var $target = document.getElementById("target");
        $target.classList.add("is-active");
    });

    $(".modal-close.is-large").click(function () {
        var $target = document.getElementById("target");
        $target.classList.remove("is-active");
    });

    $(".open-node").click(function () {
        var row = this.parentNode.parentNode.parentNode;
        var lv = row.className.substring(3) * 1;
        if (row.querySelector("td input").checked) {
            HideLevel(row, lv);
        } else {
            ShowLevel(row, lv);
        }
    });

    function ShowLevel(row, lv) {
        var tBody = row.parentNode;
        var i = row.rowIndex;
        row = tBody.rows[i];
        while (row && row.className.substring(3) * 1 > lv) {
            if (row.className.substring(3) * 1 == lv + 1) {
                row.style.display = "table-row";
                if (
                    row.querySelector("td input") &&
                    row.querySelector("td input").checked
                ) {
                    ShowLevel(row, lv + 1);
                }
            }
            i += 1;
            row = tBody.rows[i];
        }
    }

    function HideLevel(row, lv) {
        var i = row.rowIndex;
        var tBody = row.parentNode;
        row = tBody.rows[i];
        while (row && row.className.substring(3) * 1 > lv) {
            row.style.display = "none";
            i += 1;
            row = tBody.rows[i];
        }
    }
});
