$(document).ready(function() {
  var opts = {
    dataType: 'json',
    success:  uploadResult,
    error: showError
  };
  $('#uploadForm').submit(function() {
    $(this).ajaxSubmit(opts);
  })
})

function uploadResult(data) {
  if (data.error) {
    $('#uploadForm').append( '<p class="error">error: "' + data.error + '"</p>' );
  }
  else if (data.id) {
    var currentUrl = window.location.href
    window.location.replace(currentUrl + data.id);
  }
}
function showError() {
  $('#uploadForm').append( '<p class="error">error: "500: Internal Server Error"</p>' );
}
