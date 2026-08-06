package com.personal.controller;

import com.personal.entity.BuildEntity;
import com.personal.repository.BuildRepository;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/builds")
public class BuildController {

    private final BuildRepository repo;

    public BuildController(BuildRepository repo) { this.repo = repo; }

    @GetMapping
    List<String> getByStatus(@RequestParam String status) {
        return repo.findByStatus(status).stream().map(BuildEntity::getId).toList();
    }

    @GetMapping("/{id}")
    BuildEntity getById(@PathVariable String id) { return repo.findById(id).orElseThrow(); }

    @PatchMapping("/{id}/status")
    BuildEntity updateStatus(@PathVariable String id, @RequestParam String status) {
        BuildEntity b = repo.findById(id).orElseThrow();
        b.setStatus(status);
        return repo.save(b);
    }
}
