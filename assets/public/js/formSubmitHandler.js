$(document).ready(function() {
  var opts = {
    dataType: 'json',
    success:  uploadResult,
    error: showError
  };
  $('.uploadForm').submit(function(e) {
    e.preventDefault();
    $(this).ajaxSubmit(opts);
  })
})
function uploadResult(data) {
  if (data.error) {
    $('.uploadForm').append( '<p class="error">error: "' + data.error + '"</p>' );
    return;
  }
  redirectToImg(data);
}
function showError() {
  $('.uploadForm').append( '<p class="error">error: "500: Internal Server Error"</p>' );
}

function redirectToImg(data) {
  var currentUrl = window.location.href;
  window.location.replace(currentUrl + data.id);
}
