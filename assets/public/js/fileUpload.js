$(document).ready(function() {
  $('.uploadFile').click(function() {
    $('.uploadFile').bind('change', function () {
      var file = $('.uploadFile')[0].files[0];
      if (/^\s*$/.test(file.name)) {
        $('.fileName').text("...");
      }
      else {
        $('.fileName').text(file.name);
      }
    });
  });
});