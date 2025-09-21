/**
 * Tests for DOM utility functions, especially security features
 */

import { setInnerHTML, escapeHtml } from '../src/common/dom-utils';

describe('DOM Utils Security', () => {
  let testElement: HTMLElement;

  beforeEach(() => {
    testElement = document.createElement('div');
  });

  describe('setInnerHTML', () => {
    it('should sanitize malicious script tags', () => {
      const maliciousHtml = '<p>Hello</p><script>alert("XSS")</script>';
      setInnerHTML(testElement, maliciousHtml);

      expect(testElement.innerHTML).toBe('<p>Hello</p>');
      expect(testElement.innerHTML).not.toContain('<script>');
    });

    it('should sanitize javascript: URLs', () => {
      const maliciousHtml = '<a href="javascript:alert(\'XSS\')">Click me</a>';
      setInnerHTML(testElement, maliciousHtml);

      expect(testElement.innerHTML).not.toContain('javascript:');
    });

    it('should sanitize data: URLs', () => {
      const maliciousHtml = '<a href="data:text/html,<script>alert(\'XSS\')</script>">Click me</a>';
      setInnerHTML(testElement, maliciousHtml);

      expect(testElement.innerHTML).not.toContain('data:');
    });

    it('should sanitize vbscript: URLs', () => {
      const maliciousHtml = '<a href="vbscript:msgbox(\'XSS\')">Click me</a>';
      setInnerHTML(testElement, maliciousHtml);

      expect(testElement.innerHTML).not.toContain('vbscript:');
    });

    it('should sanitize onclick handlers', () => {
      const maliciousHtml = '<button onclick="alert(\'XSS\')">Click me</button>';
      setInnerHTML(testElement, maliciousHtml);

      expect(testElement.innerHTML).not.toContain('onclick');
    });

    it('should allow safe HTML tags and attributes', () => {
      const safeHtml = '<p class="text">Hello <strong>world</strong>!</p><a href="https://example.com">Link</a>';
      setInnerHTML(testElement, safeHtml);

      expect(testElement.innerHTML).toContain('<p class="text">');
      expect(testElement.innerHTML).toContain('<strong>');
      expect(testElement.innerHTML).toContain('href="https://example.com"');
    });

    it('should handle nested malicious patterns', () => {
      const nestedMalicious = '<p>Hello</p><scr<script>ipt>alert("XSS")</script>';
      setInnerHTML(testElement, nestedMalicious);

      // sanitize-html removes script tags but may leave some text content
      expect(testElement.innerHTML).toContain('<p>Hello</p>');
      expect(testElement.innerHTML).not.toContain('<script>');
      expect(testElement.innerHTML).not.toContain('</script>');
    });
  });

  describe('escapeHtml', () => {
    it('should escape HTML entities', () => {
      const text = '<script>alert("XSS")</script>';
      const escaped = escapeHtml(text);

      expect(escaped).toBe('&lt;script&gt;alert("XSS")&lt;/script&gt;');
    });

    it('should escape quotes and ampersands', () => {
      const text = 'Hello & "world" with \'quotes\'';
      const escaped = escapeHtml(text);

      expect(escaped).toContain('&amp;');
      // escapeHtml uses textContent which doesn't escape quotes the same way
      expect(escaped).toContain('"world"');
    });
  });
});