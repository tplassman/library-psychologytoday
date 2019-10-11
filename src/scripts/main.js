import './polyfills';
import pop from 'compop';

// Components
import BookForm from './components/book-form';

const { SITE_HANDLE } = process.env;

/* eslint-disable quote-props */
const classMap = {
	'book-form': BookForm,
};
/* eslint-enable quote-props */

const actions = {
	// Action events
	lockScroll: 'lock-scroll',
	unlockScroll: 'unlock-scroll',
	showHeader: 'show-header',
	hideHeader: 'hide-header',
	closeHeader: 'close-header',
};


// Event handler functions
function handleDOMConentLoaded() {
    const scaffold = window[SITE_HANDLE];

    function cb() {
        // Do something after components initialize
    }

    // Call component constructors
    pop({ scaffold, classMap, actions, cb });
}

// Add event listeners
document.addEventListener('DOMContentLoaded', handleDOMConentLoaded);
