$(document).ready(function () {

    $(".navbar-burger").click(function () {
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    });

    $(".open-reciept").click(function () {

        var $target = document.getElementById("target");
        var $type = "debits";
        if (this.getAttribute("debits") == null) {
            $type = "credits"
        }

        $.ajax({
            url: encodeURI("./receipt"),
            method: "post",
            dataType: "json",
            data: JSON.stringify({
                type: $type,
                id: parseInt(this.getAttribute($type)),
            }),
            success: function (data) {
                $target.querySelector("#sum").innerHTML = data.Sum.toLocaleString() + " ₽";
                $target.querySelector("#tag").innerHTML = data.Name;
                $target.querySelector("#username").innerHTML = data.FullName;
                $target.querySelector("#userpic").setAttribute("src", "https://bulma.io/images/placeholders/96x96.png");
                if (data.UserPic != "") {
                    $target.querySelector("#userpic").setAttribute("src", data.UserPic);
                }
                $target.querySelector("#receipt").setAttribute("src", "https://bulma.io/images/placeholders/1280x960.png");
                if (data.Receipt != "") {
                    $target.querySelector("#receipt").setAttribute("src", data.Receipt);
                }
                $target.querySelector("#comment").innerHTML = data.Comment;
                $target.classList.add("is-active");
            }
        });

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
