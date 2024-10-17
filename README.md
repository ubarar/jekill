# jekill

Built is a part drop in replacement for `jekyll`. Serves the currently directory as a website, rendering markdown files with common extensions.

### Why?

- Jekyll is very hard to run locally, as I discovered in [my attempt to do so](https://blog.sadboi.dev/projects/running-jekyll)
	- I need a fast way to preview my website locally, and local markdown previewers don't render relative links to media correctly
	- Some of my pages don't use Markdown at all, so I need to just serve that file directly
	- Jekyll builds your site and serves it. This means that I can't just mount my assets into a container and then serve it. This makes it that much harder to deploy the website locally or otherwise. My website is already on the order of 10GB - and I expect this to keep growing
- Other Markdown website generators like [Hugo](https://gohugo.io/documentation/) impose too much structure. I wasn't willing to migrate my posts for the sake of using a specific tool

### How

Once build, simply run with `jekill`, adding flags `-addr` and `-port` as needed. To add stylesheets to your markdown pages, ensure that there is a `.config/head.html` from where you start `jekill`. The contents of this file will become the innerHTML for the head of every markdown page.

