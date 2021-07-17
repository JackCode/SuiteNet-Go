// Add triangle to mark which page is active
var navLinks = document.querySelectorAll("nav a");

//window.location.pathname.split('/')[1].toLowerCase()
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href').includes(window.location.pathname.split('/')[1].toLowerCase())) {
		link.classList.add("live");
		break;
	}
}

// Get the modal
var modal = document.getElementById("addNoteModal");

// Get the button that opens the modal
var btn = document.getElementById("addNoteBtn");

// Get the <span> element that closes the modal
var span = document.getElementsByClassName("close")[0];

// When the user clicks on the button, open the modal
btn.onclick = function() {
  modal.style.display = "block";
}

// When the user clicks on <span> (x), close the modal
span.onclick = function() {
  modal.style.display = "none";
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
  if (event.target == modal) {
    modal.style.display = "none";
  }
}

