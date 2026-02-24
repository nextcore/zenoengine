# Installation

## Meet ZenoEngine

ZenoEngine is a web application framework with expressive, elegant syntax. We've already laid the foundation — freeing you to create without sweating the small things.

If you are coming from Laravel, you will feel right at home. ZenoEngine brings the elegant syntax, routing, ORM, and Blade templating you love, but executes them at the speed of compiled Go.

## Installing ZenoEngine

ZenoEngine is distributed as a single compiled binary. There is no need to install PHP, Composer, or Nginx.

### Using the Zeno CLI

You can create a new ZenoEngine project instantly using the CLI:

```bash
zeno new blog
cd blog
zeno dev
```

That's it! Your application is now running on `http://localhost:3000`.

## Directory Structure

When you create a new ZenoEngine application, the directory structure is designed to be familiar:

```text
app/
  controllers/
  models/
bootstrap/
config/
database/
  migrations/
  seeders/
public/
resources/
  views/
  css/
routes/
  web.zl
  api.zl
```
