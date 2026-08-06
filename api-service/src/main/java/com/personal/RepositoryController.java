package com.personal;

import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/repositories")
public class RepositoryController {

    private final RepositoryRepository repo;

    public RepositoryController(RepositoryRepository repo) { this.repo = repo; }

    @GetMapping
    List<RepositoryEntity> getAll() { return repo.findAll(); }

    @GetMapping("/{id}")
    RepositoryEntity getById(@PathVariable String id) { return repo.findById(id).orElseThrow(); }

    @PostMapping
    RepositoryEntity create(@RequestBody RepositoryEntity r) { return repo.save(r); }

    @PutMapping("/{id}")
    RepositoryEntity update(@PathVariable String id, @RequestBody RepositoryEntity r) {
        RepositoryEntity existing = repo.findById(id).orElseThrow();
        existing.setName(r.getName());
        existing.setUrl(r.getUrl());
        return repo.save(existing);
    }

    @DeleteMapping("/{id}")
    void delete(@PathVariable String id) { repo.deleteById(id); }
}
