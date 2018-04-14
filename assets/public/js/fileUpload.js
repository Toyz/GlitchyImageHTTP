$('#uploadFile').click(function() {
  $('#uploadFile').bind('change', function () {
    var file = $("#uploadFile")[0].files[0];
    console.log(file.name);
    if (/^\s*$/.test(file.name)) {
      $(".fileName").text("...");
    }
    else {
      $(".fileName").text(file.name);
    }
  });
});
