export default class {
	constructor({
		id,
		formHandle,
		searchIsbn = false,
	}) {
		const el = document.getElementById(id);
		const form = el.querySelector(formHandle || 'form');
		const isbn = form.querySelector('[name="isbn"]');
		const limited = form.querySelectorAll('[maxlength]');
		console.log(form);
		console.log(limited);

		function handleIsbn(e) {
			console.log(`TODO: Query ISBN API for: ${isbn.value}, and prepopulate results`);
		}
		const handleLimitCounts = Array.from(limited).map(l => {
			// Create count element
			const em = document.createElement('em');
			const maxLength = l.getAttribute('maxlength');

			l.parentElement.appendChild(em);

			return function() {
				em.textContent = `${maxLength - l.value.length}`;
			}
		});

		if (searchIsbn) {
			isbn.addEventListener('keyup', handleIsbn);
		}
		limited.forEach((l, i) => {
			l.addEventListener('keyup', handleLimitCounts[i]);
		});
	}
}
