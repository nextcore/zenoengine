package cli

import (
	"fmt"
	"os"
)

func HandleMakeAuth() {
	fmt.Println("🛠️  Scaffolding Authentication System...")

	// 1. Create Directories
	dirs := []string{
		"src/modules/auth",
		"views/auth",
		"migrations",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Printf("❌ Failed to create directory %s: %v\n", d, err)
			return
		}
	}

	// 2. (Skipped) Migration is now in 001_init_schema.zl

	// 3. Generate Views
	createFileIfNotExists("views/auth/login.html", loginViewTemplate)
	createFileIfNotExists("views/auth/register.html", registerViewTemplate)

	// 4. Generate Modules
	createFileIfNotExists("src/modules/auth/login.zl", loginModuleTemplate)
	createFileIfNotExists("src/modules/auth/register.zl", registerModuleTemplate)
	createFileIfNotExists("src/modules/auth/logout.zl", logoutModuleTemplate)

	fmt.Println("\n✅ Auth Scaffolding Generated Successfully!")
	fmt.Println("\n👉 Next Steps:")
	fmt.Println("1. Run migration:  go run ./cmd/zeno/zeno.go migrate")
	fmt.Println("2. Register routes in 'src/main.zl':")
	fmt.Println(`
   http.group: /auth {
     do: {
       include: src/modules/auth/login.zl
       include: src/modules/auth/register.zl
       include: src/modules/auth/logout.zl
     }
   }`)
	os.Exit(0)
}

func createFileIfNotExists(path string, content string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("⚠️  File already exists, skipping: %s\n", path)
		return
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("❌ Failed to create file %s: %v\n", path, err)
	} else {
		fmt.Printf("✅ Created: %s\n", path)
	}
}

// --- TEMPLATES ---
const loginViewTemplate = `
{{ define "content" }}
<div class="row justify-content-center mt-5">
    <div class="col-md-5">
        <div class="card shadow-sm border-0">
            <div class="card-body p-4">
                <h3 class="fw-bold text-center mb-4">Login</h3>
                
                {{ if .error }}
                <div class="alert alert-danger">{{ .error }}</div>
                {{ end }}

                <form action="/auth/login" method="POST">
                    <input type="hidden" name="gorilla.csrf.Token" value="{{ .csrf }}">
                    
                    <div class="mb-3">
                        <label class="form-label text-muted small fw-bold">Email Address</label>
                        <input type="email" name="email" class="form-control" required placeholder="name@example.com">
                    </div>

                    <div class="mb-4">
                        <label class="form-label text-muted small fw-bold">Password</label>
                        <input type="password" name="password" class="form-control" required>
                    </div>

                    <button type="submit" class="btn btn-primary w-100 py-2 fw-bold">Sign In</button>
                    
                    <div class="text-center mt-3">
                        <small class="text-muted">Don't have an account? <a href="/auth/register" class="text-decoration-none">Register</a></small>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
{{ end }}
`

const registerViewTemplate = `
{{ define "content" }}
<div class="row justify-content-center mt-5">
    <div class="col-md-5">
        <div class="card shadow-sm border-0">
            <div class="card-body p-4">
                <h3 class="fw-bold text-center mb-4">Register</h3>

                {{ if .error }}
                <div class="alert alert-danger">{{ .error }}</div>
                {{ end }}

                <form action="/auth/register" method="POST">
                    <input type="hidden" name="gorilla.csrf.Token" value="{{ .csrf }}">
                    
                    <div class="mb-3">
                        <label class="form-label text-muted small fw-bold">Full Name</label>
                        <input type="text" name="name" class="form-control" required placeholder="John Doe">
                    </div>

                    <div class="mb-3">
                        <label class="form-label text-muted small fw-bold">Email Address</label>
                        <input type="email" name="email" class="form-control" required placeholder="name@example.com">
                    </div>

                    <div class="mb-4">
                        <label class="form-label text-muted small fw-bold">Password</label>
                        <input type="password" name="password" class="form-control" required>
                    </div>

                    <button type="submit" class="btn btn-primary w-100 py-2 fw-bold">Create Account</button>
                    
                    <div class="text-center mt-3">
                        <small class="text-muted">Already registered? <a href="/auth/login" class="text-decoration-none">Login</a></small>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
{{ end }}
`

const loginModuleTemplate = `
http.get: /login {
    do: {
        sec.csrf_token: $csrf
        view.render: auth/login.html {
            title: "Login"
            csrf: $csrf
        }
    }
}

http.post: /login {
    do: {
        http.form: email { as: $email }
        http.form: password { as: $password }

        try: {
            do: {
                auth.login: {
                    email: $email
                    password: $password
                    as: $token
                }
                
                http.cookie: set {
                    name: "token"
                    value: $token
                    http_only: true
                    path: "/"
                    max_age: 259200 // 3 days
                }

                http.redirect: /
            }
            catch: {
                sec.csrf_token: $csrf
                view.render: auth/login.html {
                    title: "Login"
                    csrf: $csrf
                    error: "Invalid email or password"
                }
            }
        }
    }
}
`

const registerModuleTemplate = `
http.get: /register {
    do: {
        sec.csrf_token: $csrf
        view.render: auth/register.html {
            title: "Register"
            csrf: $csrf
        }
    }
}

http.post: /register {
    do: {
        http.form: name { as: $name }
        http.form: email { as: $email }
        http.form: password { as: $password }

        if: $password != "" {
            then: {
                // Hash Password
                crypto.hash: $password { as: $hashed_pass }

                try: {
                    do: {
                        db.table: users
                        db.insert: {
                            name: $name
                            email: $email
                            password: $hashed_pass
                        }
                        http.redirect: /auth/login
                    }
                    catch: {
                        sec.csrf_token: $csrf
                        view.render: auth/register.html {
                            title: "Register"
                            csrf: $csrf
                            error: "Registration failed. Email might be taken."
                        }
                    }
                }
            }
            else: {
                 sec.csrf_token: $csrf
                 view.render: auth/register.html {
                     title: "Register"
                     csrf: $csrf
                     error: "Password is required"
                 }
            }
        }
    }
}
`

const logoutModuleTemplate = `
http.post: /logout {
    do: {
        http.cookie: delete {
            name: "token"
            path: "/"
        }
        http.redirect: /auth/login
    }
}

http.get: /logout {
    do: {
        http.cookie: delete {
            name: "token"
            path: "/"
        }
        http.redirect: /auth/login
    }
}
`
