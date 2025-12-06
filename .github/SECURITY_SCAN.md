# Security Scanning

This document describes the security scanning processes implemented in the Fox project.

## Overview

Fox implements multiple layers of security scanning to ensure code quality and protect against vulnerabilities:

1. **Vulnerability Scanning (govulncheck)** - Detects known vulnerabilities in dependencies
2. **Static Application Security Testing (CodeQL)** - Analyzes code for security issues
3. **Dependency Review** - Reviews dependency changes in pull requests

## Automated Security Scans

### Schedule

- **On every push** to main branch
- **On every pull request** to main branch
- **Weekly schedule**: Every Monday at 00:00 UTC

### Scan Types

#### 1. govulncheck - Go Vulnerability Database

**What it does:**
- Scans all Go dependencies against the official Go vulnerability database
- Checks both direct and indirect dependencies
- Identifies vulnerable code paths in your application

**How to run locally:**
```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run vulnerability scan
govulncheck ./...

# Verbose output
govulncheck -show verbose ./...
```

**Configuration:**
- Workflow: `.github/workflows/security.yml`
- Job: `govulncheck`
- Documentation: https://go.dev/security/vuln/

#### 2. CodeQL - Static Analysis Security Testing (SAST)

**What it does:**
- Performs deep semantic analysis of Go code
- Detects security vulnerabilities like:
  - SQL injection
  - Command injection
  - Path traversal
  - Cross-site scripting (XSS)
  - Authentication/authorization issues
  - Cryptographic weaknesses
- Uses security-extended and security-and-quality query suites

**Configuration:**
- Workflow: `.github/workflows/security.yml`
- Job: `codeql`
- Query suites: `security-extended`, `security-and-quality`
- Documentation: https://codeql.github.com/

**Viewing results:**
- Navigate to: `Security` → `Code scanning alerts` in GitHub repository

#### 3. Dependency Review

**What it does:**
- Reviews dependency changes in pull requests
- Checks for:
  - Known vulnerabilities in new dependencies
  - License compliance issues
  - Supply chain risks

**Configuration:**
- Workflow: `.github/workflows/security.yml`
- Job: `dependency-review`
- Fail on severity: `moderate` or higher
- Denied licenses: `GPL-2.0`, `GPL-3.0`

## Security Workflow Status

Current status of security checks:

| Check | Status | Description |
|-------|--------|-------------|
| govulncheck | ✅ Enabled | Vulnerability scanning |
| CodeQL | ✅ Enabled | SAST analysis |
| Dependency Review | ✅ Enabled | PR dependency checks |
| golangci-lint | ✅ Enabled | Code quality & security linting |

## Interpreting Results

### govulncheck Results

**No vulnerabilities found:**
```
Your code is not affected by any known vulnerabilities.
```

**Vulnerabilities found:**
```
Vulnerability #1: GO-YYYY-NNNN
    Description of the vulnerability
  More info: https://pkg.go.dev/vuln/GO-YYYY-NNNN
  Module: example.com/vulnerable-package
    Found in: example.com/vulnerable-package@v1.2.3
    Fixed in: example.com/vulnerable-package@v1.2.4
```

**Action required:**
1. Review the vulnerability details at the provided link
2. Update the affected package: `go get -u example.com/vulnerable-package@v1.2.4`
3. Run `go mod tidy`
4. Re-run tests and security scans

### CodeQL Results

CodeQL results are available in the GitHub Security tab:
1. Go to `Security` → `Code scanning`
2. Review any alerts
3. Click on an alert for detailed information
4. Fix the issue and create a pull request

**Alert severity levels:**
- **Critical**: Must fix immediately
- **High**: Should fix in current sprint
- **Medium**: Should fix soon
- **Low**: Fix when convenient
- **Note**: Informational, consider fixing

### Standard Library Vulnerabilities

If govulncheck reports vulnerabilities in Go's standard library:

```
Found in: crypto/x509@go1.25.4
Fixed in: crypto/x509@go1.25.5
```

**Action required:**
1. Upgrade Go to the fixed version
2. Update `go.mod` if it specifies a Go version
3. Update CI/CD workflows if they pin a specific version
4. Re-run security scans

## Best Practices

### For Contributors

1. **Before submitting a PR:**
   ```bash
   # Run security checks locally
   govulncheck ./...
   golangci-lint run
   ```

2. **Fix security issues immediately**
   - Security issues block merges
   - Get help from maintainers if needed

3. **Keep dependencies updated**
   - Regularly update dependencies
   - Monitor security advisories

### For Maintainers

1. **Review security alerts weekly**
   - Check GitHub Security tab
   - Address high/critical issues immediately

2. **Update dependencies regularly**
   ```bash
   # Check for updates
   go list -u -m all

   # Update dependencies
   go get -u ./...
   go mod tidy
   ```

3. **Monitor vulnerability databases**
   - Go vulnerability database: https://vuln.go.dev/
   - GitHub Security Advisories
   - Dependabot alerts

## Continuous Improvement

The security scanning configuration is continuously improved:

- ✅ **Completed (2025-12-06)**
  - Integrated govulncheck for vulnerability scanning
  - Added CodeQL SAST analysis
  - Configured dependency review for PRs
  - Set up weekly automated scans

- 📋 **Future enhancements**
  - Add Snyk or similar third-party scanning
  - Implement container image scanning
  - Add security scorecards
  - Set up automated security updates

## Reporting Security Issues

If you discover a security vulnerability, please refer to [SECURITY.md](../SECURITY.md) for our responsible disclosure process.

**Do not** open public GitHub issues for security vulnerabilities.

## Resources

- [Go Security Policy](https://go.dev/security/policy)
- [Go Vulnerability Database](https://vuln.go.dev/)
- [CodeQL Documentation](https://codeql.github.com/docs/)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

## Questions?

For questions about security scanning:
1. Check this documentation
2. Review [SECURITY.md](../SECURITY.md)
3. Open a discussion in GitHub Discussions
4. Contact maintainers

---

**Last Updated**: 2025-12-06
