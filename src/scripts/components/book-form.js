export default class {
	constructor({
		id,
	}) {
		const el = document.getElementById(id);
		const form = el.querySelector('form');
		const isbnInput = form.querySelector('[name="isbn"]');

		function handleIsbnInput(e) {
			console.log(`TODO: Query ISBN API for: ${isbnInput.value}, and prepopulate results`);
		}

		isbnInput.addEventListener('keyup', handleIsbnInput);
	}
}
