package com.personal.controllers;

import com.personal.entities.JobEntity;
import com.personal.repositories.JobRepository;
import com.personal.repositories.RepositoryRepository;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/jobs")
public class JobController {

    private final JobRepository repo;
    private final RepositoryRepository repositoryRepo;

    public JobController(JobRepository repo, RepositoryRepository repositoryRepo) {
        this.repo = repo;
        this.repositoryRepo = repositoryRepo;
    }

    @GetMapping
    List<JobEntity> getAll() { return repo.findAll(); }

    @GetMapping("/{id}")
    JobEntity getById(@PathVariable String id) { return repo.findById(id).orElseThrow(); }

    @PostMapping
    JobEntity create(@RequestBody JobRequest request) {
        JobEntity j = new JobEntity();
        j.setName(request.name());
        j.setRepository(repositoryRepo.findById(request.repositoryId()).orElseThrow());
        return repo.save(j);
    }

    @DeleteMapping("/{id}")
    void delete(@PathVariable String id) { repo.deleteById(id); }

    record JobRequest(String name, String repositoryId) {}
}
