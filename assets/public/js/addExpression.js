$(document).ready(function() {
  var counter = 1
  var limit = 5
  $('#addExpr').click(function() {
    if (counter == limit) {
      alert("You have reached the limit of adding " + counter + " expressions");
    }
    else {
      counter++;
      $('#expressionList').append( '<div class="expressionForm" id=' + (counter + 1) + '><label for="expression"><span>>> expression.( " </span></label><input type="text" name="expression" value="" class="expression" /><span> " )</span></div>' );
      if (counter > 1) {
        $('#removeExpr').addClass('unhide');
      }
    }
  })
  $('#removeExpr').click(function() {
    $('.expressionForm#'+counter).remove();
    counter--;
    if (counter == 1) {
      $('#removeExpr').removeClass('unhide');
    }
  })
})
