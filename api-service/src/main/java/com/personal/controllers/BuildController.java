package com.personal.controllers;

import com.personal.entities.BuildEntity;
import com.personal.entities.JobEntity;
import com.personal.repositories.BuildRepository;
import com.personal.repositories.JobRepository;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.List;

@RestController
@RequestMapping("/api/builds")
public class BuildController {

    private final BuildRepository repo;
    private final JobRepository jobRepo;

    public BuildController(BuildRepository repo, JobRepository jobRepo) {
        this.repo = repo;
        this.jobRepo = jobRepo;
    }

    @GetMapping
    List<BuildEntity> getAll() { return repo.findAll(); }

    @GetMapping("/{id}")
    BuildEntity getById(@PathVariable String id) { return repo.findById(id).orElseThrow(); }

    @GetMapping("/jobs/{jobId}")
    List<BuildEntity> getByJob(@PathVariable String jobId) { return repo.findByJobId(jobId); }

    @PostMapping("/jobs/{jobId}")
    BuildEntity build(@PathVariable String jobId) {
        JobEntity job = jobRepo.findById(jobId).orElseThrow();
        BuildEntity b = new BuildEntity();
        b.setJob(job);
        b.setRepositoryUrl(job.getRepository().getUrl());
        b.setBuildNumber(repo.countByJobId(jobId) + 1);
        b.setTimestamp(Instant.now());
        b.setStatus("QUEUED");
        return repo.save(b);
    }

    @DeleteMapping("/{id}")
    void delete(@PathVariable String id) { repo.deleteById(id); }
}
