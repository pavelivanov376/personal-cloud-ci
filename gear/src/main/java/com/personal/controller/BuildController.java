package com.personal.controller;

import com.personal.entity.Build;
import com.personal.repository.BuildRepository;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/builds")
public class BuildController {

    private final BuildRepository repo;

    public BuildController(BuildRepository repo) { this.repo = repo; }

    @GetMapping
    List<Build> getByStatus(@RequestParam String status) { return repo.findByStatus(status); }
}
