/**
 * gulpfile.js
 */

/**
 * CONFIG
 *
 * Require packages
 */
const {
	src,
	dest,
	watch,
	task,
	series,
	parallel,
} = require('gulp'),
	autoprefixer = require('gulp-autoprefixer'),
	babel = require('babelify'),
	browserify = require('browserify'),
	buffer = require('vinyl-buffer'),
	cheerio = require('gulp-cheerio'),
	del = require('del'),
	eslint = require('gulp-eslint'),
	env = require('dotenv'),
	envify = require('envify'),
	gulpif = require('gulp-if'),
	imagemin = require('gulp-imagemin'),
	livereload = require('gulp-livereload'),
	minify = require('gulp-clean-css'),
	replace = require('gulp-string-replace'),
	sass = require('gulp-sass'),
	source = require('vinyl-source-stream'),
	svgstore = require('gulp-svgstore'),
	stylelint = require('gulp-stylelint'),
	uglify = require('gulp-uglify'),
	watchify = require('watchify');

/**
 * ENV
 *
 * Set user environment variables.
 * Require statements need to be declared before this setting.
 */
env.config();

/**
 * ROOTS
 */
const srcroot = 'src';
const dstroot = 'static';

/**
 * CONFIG
 *
 * Create variables
 */
const config = {
	images: {
		src: './' + srcroot + '/images/',
		dst: './' + dstroot + '/images/',
	},

	scripts: {
		entry: './' + srcroot + '/scripts/main.js',
		src: './' + srcroot + '/scripts/**/*.js',
		dst: './' + dstroot + '/scripts/',
	},

	styles: {
		entry: './' + srcroot + '/styles/main.scss',
		src: './' + srcroot + '/styles/**/*.scss',
		dst: './' + dstroot + '/styles/',
	},
};

/**
 * CLEAN
 *
 * Individual clean tasks
 */
task('clean:images', cb => {
	del([config.images.dst]);
	cb();
});
task('clean:scripts', cb => {
	del([config.scripts.dst]);
	cb();
});
task('clean:styles', cb => {
	del([config.styles.dst]);
	cb();
});

/**
 * IMAGES:RASTER
 *
 * Copies raster images to configured dist directory and
 * performs minimal optimization.
 *
 * optimizationLevel set to 0 to disable CPU-intensive image crunching.
 * Use ImageOptim (lossless) on your source images.
 *
 * We do want images to be progressive and interlaced, though.
 */
task('images:raster', () => src([
	config.images.src + '**/*.gif',
	config.images.src + '**/*.jpg',
	config.images.src + '**/*.png',
	config.images.src + '**/*.ico',
])
	.pipe(imagemin([
		imagemin.gifsicle({
			interlaced: true,
			optimizationLevel: 1
		}),
		imagemin.jpegtran({
			progressive: true
		}),
	]))
	.pipe(dest(config.images.dst))
);

/**
 * IMAGES:VECTOR:SPRITES
 *
 * Combine all svgs in target directory into a single spritemap.
 */
task('images:vector:sprites', () => src([
	config.images.src + 'sprites/*.svg',
])
	.pipe(svgstore({ inlineSvg: true }))
	.pipe(cheerio({
		run: $ => {
			$('svg').attr('style', 'display:none'); // make sure the spritemap doesn't show
			$('[fill]').removeAttr('fill'); // remove all 'fill' attributes in order to control via CSS
		},
		parserOptions: { lowerCaseAttributeNames: false },
	}))
	.pipe(dest(config.images.dst + 'sprites/'))
);

/**
 * IMAGES:VECTOR:PASSTHROUGH
 *
 * Regardless of other processing, at least copy all vectors to dist
 */
task('images:vector:passthrough', () => src([
	config.images.src + '**/*.svg',
])
	.pipe(dest(config.images.dst))
);

/**
 * SCRIPTS:FORMAT
 *
 * Autofix stylistic syntax in your source files using .eslintrc
 */
task('scripts:format', () => src([
	config.scripts.src,
])
	.pipe(eslint({
		fix: true,
	}))
	.pipe(eslint.format())
	.pipe(gulpif(file => (
		file.eslint != null && file.eslint.fixed
	), dest(config.scripts.dst.replace(dstroot, srcroot))))
);

/**
 * SCRIPTS:BROWSERIFY
 *
 * Transpile and bundle JS
 */
task('scripts:browserify', () => {
	const bundler = browserify(config.scripts.entry)
		.transform(babel)
		.transform(envify);

	return bundler.bundle()
		.on('error', err => { console.error(err); })
		.pipe(source('main.js'))
		.pipe(buffer())
		.pipe(uglify())
		.pipe(dest(config.scripts.dst));
});

/**
 * SCRIPTS:WATCH
 *
 * Watch JS for changes
 */
task('scripts:watch', () => {
	livereload.listen();

	const bundler = watchify(browserify(config.scripts.entry, { debug: true })
		.transform(babel)
		.transform(envify));

	bundler.on('update', rebundle);

	function rebundle() {
		return bundler.bundle()
			.on('error', err => { console.error(err); })
			.pipe(source('main.js'))
			.pipe(buffer())
			.pipe(uglify())
			.pipe(dest(config.scripts.dst))
			.pipe(livereload());
	}

	return rebundle();
});

/**
 * STYLES:SASS
 *
 * Compile SASS
 */
task('styles:sass', () => src([
	config.styles.entry,
])
	.pipe(sass().on('error', err => {
		console.error(err);
	}))
	.pipe(autoprefixer({
		cascade: false,
	}))
	.pipe(minify({
		mediaMerging: true,
		processImport: true,
		roundingPrecision: 10,
	}))
	.pipe(dest(config.styles.dst))
	.pipe(livereload())
);

/**
 * STYLES:WATCH
 *
 * Watch SASS for changes
 */
task('styles:watch', () => {
	livereload.listen();
	watch(config.styles.src, series(['styles:sass']));
});

/**
 * STYLES:FORMAT
 *
 * Autofix stylistic syntax in your source files using .stylelintrc
 */
task('styles:format', () => src([
	config.styles.src,
])
	.pipe(stylelint({
		fix: true,
		reporters: [
			{ formatter: 'string', console: true },
		],
	}))
	.pipe(gulpif(file => {
		// Hack: if the file was skipped by .stylelintignore file content will be []
		// gulp-stylelint doesn't seems to catch this and rewrites the file to []
		return file.contents.length > 2;
	}, dest(config.styles.dst.replace(dstroot, srcroot))))
);

/**
 * WATCH
 *
 * Watch styles and scripts
 */
task('watch', parallel(['styles:watch', 'scripts:watch']));

/**
 * TASK WRAPPERS
 *
 * Wrappers for the individual build tasks as well as a bundled build task
 */
task('clean', parallel(['clean:images', 'clean:scripts', 'clean:styles']));
task('images', parallel(['images:raster', 'images:vector:sprites', 'images:vector:passthrough']));
task('scripts', parallel(['scripts:browserify']));
task('styles', parallel(['styles:sass']));
task('build', parallel(['images', 'scripts', 'styles']));

/**
 * DEFAULT
 *
 * This list will be written to the terminal when the default Gulp task is run.
 * The intention is to use direct tasks instead of a vague reference to the default task.
 */
task('default', cb => {
	console.log('\nHello!\n\nThis gulpfile doesn\'t do anything by default. Please use the following to see a list of available tasks:\n\n$ gulp --tasks-simple\n\n');
	cb();
});
