package parser

// HTMLFixtures contains sample HTML content for testing
type HTMLFixtures struct {
	Name        string
	BaseURL     string
	HTMLContent string
	Expected    []string
	Description string
}

// GetTestFixtures returns a comprehensive set of HTML test fixtures
func GetTestFixtures() []HTMLFixtures {
	return []HTMLFixtures{
		{
			Name:    "Simple links",
			BaseURL: "https://example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
	<a href="/page1">Page 1</a>
	<a href="/page2">Page 2</a>
	<a href="https://other.com/external">External</a>
</body>
</html>`,
			Expected: []string{
				"https://example.com/page1",
				"https://example.com/page2",
				"https://other.com/external",
			},
			Description: "Basic HTML with relative and absolute links",
		},
		{
			Name:    "Complex navigation",
			BaseURL: "https://blog.example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Blog</title></head>
<body>
	<nav>
		<a href="/">Home</a>
		<a href="/about">About</a>
		<a href="/posts">Posts</a>
		<a href="../admin">Admin</a>
	</nav>
	<main>
		<article>
			<h1><a href="/posts/2023/article-1">Article 1</a></h1>
			<p>Content with <a href="/posts/2023/article-2">another article</a></p>
		</article>
	</main>
	<footer>
		<a href="mailto:contact@example.com">Contact</a>
		<a href="tel:+1234567890">Call</a>
		<a href="#top">Back to top</a>
	</footer>
</body>
</html>`,
			Expected: []string{
				"https://blog.example.com/",
				"https://blog.example.com/about",
				"https://blog.example.com/posts",
				"https://blog.example.com/admin",
				"https://blog.example.com/posts/2023/article-1",
				"https://blog.example.com/posts/2023/article-2",
			},
			Description: "Complex navigation with various link types (filtered)",
		},
		{
			Name:    "E-commerce page",
			BaseURL: "https://shop.example.com/category/electronics",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Electronics</title></head>
<body>
	<div class="products">
		<div class="product">
			<a href="./laptop-1">
				<img src="/images/laptop1.jpg" alt="Laptop 1">
				<h3>Gaming Laptop</h3>
			</a>
			<a href="./laptop-1?tab=reviews" class="reviews-link">Reviews</a>
		</div>
		<div class="product">
			<a href="/category/electronics/phone-1">Smartphone</a>
			<a href="/category/electronics/phone-1#specifications">Specs</a>
		</div>
	</div>
	<div class="pagination">
		<a href="?page=1">1</a>
		<a href="?page=2" class="current">2</a>
		<a href="?page=3">3</a>
		<a href="?page=2&sort=price">Sort by Price</a>
	</div>
</body>
</html>`,
			Expected: []string{
				"https://shop.example.com/category/laptop-1",
				"https://shop.example.com/category/laptop-1?tab=reviews",
				"https://shop.example.com/category/electronics/phone-1",
				"https://shop.example.com/category/electronics/phone-1",
				"https://shop.example.com/category/electronics?page=1",
				"https://shop.example.com/category/electronics?page=2",
				"https://shop.example.com/category/electronics?page=3",
				"https://shop.example.com/category/electronics?page=2&sort=price",
			},
			Description: "E-commerce page with products and pagination",
		},
		{
			Name:    "Social media links",
			BaseURL: "https://company.example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Company</title></head>
<body>
	<div class="social-links">
		<a href="https://twitter.com/company">Twitter</a>
		<a href="https://facebook.com/company">Facebook</a>
		<a href="https://linkedin.com/company/company">LinkedIn</a>
		<a href="https://github.com/company">GitHub</a>
	</div>
	<div class="internal-links">
		<a href="/team">Our Team</a>
		<a href="/careers">Careers</a>
		<a href="/blog/">Blog</a>
	</div>
	<div class="special-links">
		<a href="javascript:void(0)" onclick="openModal()">Open Modal</a>
		<a href="#section1">Section 1</a>
		<a href="data:text/plain;base64,SGVsbG8gV29ybGQ=">Data URL</a>
		<a href="ftp://files.example.com/doc.pdf">FTP File</a>
	</div>
</body>
</html>`,
			Expected: []string{
				"https://twitter.com/company",
				"https://facebook.com/company",
				"https://linkedin.com/company/company",
				"https://github.com/company",
				"https://company.example.com/team",
				"https://company.example.com/careers",
				"https://company.example.com/blog",
			},
			Description: "Mix of external social links and internal links (filtered)",
		},
		{
			Name:    "Malformed HTML",
			BaseURL: "https://broken.example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Broken Page</title>
<body>
	<div class="content">
		<a href="/page1">Page 1</a>
		<a href="/page2">Page 2
		<div>
			<a href="/nested/page">Nested</a>
		</div>
	</div>
</body>
</html>`,
			Expected: []string{
				"https://broken.example.com/page1",
				"https://broken.example.com/page2",
				"https://broken.example.com/page2",
				"https://broken.example.com/nested/page",
			},
			Description: "Malformed HTML that should still be parsed correctly",
		},
		{
			Name:    "Empty and edge cases",
			BaseURL: "https://edge.example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Edge Cases</title></head>
<body>
	<a href="">Empty href</a>
	<a href="   ">Whitespace href</a>
	<a href=".">Current directory</a>
	<a href="..">Parent directory</a>
	<a href="/">Root</a>
	<a href="./same-dir">Same directory</a>
	<a href="../parent-dir">Parent directory</a>
	<a href="?query=test">Query only</a>
	<a href="#fragment">Fragment only</a>
	<a href="?query=test#fragment">Query with fragment</a>
	<a>No href attribute</a>
	<a href="/normal" title="Normal link">Normal</a>
</body>
</html>`,
			Expected: []string{
				"https://edge.example.com/",
				"https://edge.example.com/",
				"https://edge.example.com/",
				"https://edge.example.com/same-dir",
				"https://edge.example.com/parent-dir",
				"https://edge.example.com/?query=test",
				"https://edge.example.com/?query=test",
				"https://edge.example.com/normal",
			},
			Description: "Edge cases with empty hrefs, relative paths, and fragments",
		},
		{
			Name:    "URL parameters and encodings",
			BaseURL: "https://api.example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>API Documentation</title></head>
<body>
	<a href="/v1/users">Users API</a>
	<a href="/v1/users?limit=10">Users with limit</a>
	<a href="/v1/users?limit=10&offset=20">Users with pagination</a>
	<a href="/v1/search?q=test+query">Search with encoded space</a>
	<a href="/v1/data?format=json&pretty=true">JSON data</a>
</body>
</html>`,
			Expected: []string{
				"https://api.example.com/v1/users",
				"https://api.example.com/v1/users?limit=10",
				"https://api.example.com/v1/users?limit=10&offset=20",
				"https://api.example.com/v1/search?q=test+query",
				"https://api.example.com/v1/data?format=json&pretty=true",
			},
			Description: "API URLs with query parameters and encodings",
		},
	}
}

// GetSameDomainTestFixtures returns fixtures for same-domain filtering tests
func GetSameDomainTestFixtures() []HTMLFixtures {
	return []HTMLFixtures{
		{
			Name:    "Mixed domain links",
			BaseURL: "https://example.com",
			HTMLContent: `<!DOCTYPE html>
<html>
<head><title>Mixed Domains</title></head>
<body>
	<a href="/internal1">Internal 1</a>
	<a href="https://example.com/internal2">Internal 2</a>
	<a href="https://sub.example.com/subdomain">Subdomain</a>
	<a href="https://other.com/external">External</a>
	<a href="https://example.org/different-tld">Different TLD</a>
	<a href="http://example.com/different-scheme">Different Scheme</a>
</body>
</html>`,
			Expected: []string{
				"https://example.com/internal1",
				"https://example.com/internal2",
				"http://example.com/different-scheme",
			},
			Description: "Mixed internal and external links for same-domain filtering",
		},
	}
}
